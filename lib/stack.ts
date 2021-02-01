import * as cdk from '@aws-cdk/core';
import * as lambda from '@aws-cdk/aws-lambda';
import * as iam from '@aws-cdk/aws-iam';
import * as sns from '@aws-cdk/aws-sns';
import * as sqs from '@aws-cdk/aws-sqs';
import * as ec2 from '@aws-cdk/aws-ec2';
import { SqsEventSource } from '@aws-cdk/aws-lambda-event-sources';
import { SqsSubscription } from '@aws-cdk/aws-sns-subscriptions';

import * as path from 'path';
import { SubnetNetworkAclAssociation, Vpc } from '@aws-cdk/aws-ec2';
import { countReset } from 'console';

export interface GitHubProps extends cdk.StackProps {
  lambdaRoleARN?: string;

  // Set either one:
  reportTopic?: sns.ITopic;
  reportTopicARN?: string;

  // A secret has `github_token`
  secretARN: string;

  // github API endpoint, default is https://api.github.com
  githubEndpoint?: string;
  // github repository to upload report.
  // e.g.) 'm-mizutani/alert' for https://github.com/m-mizutani/alert
  githubRepo: string;

  // Optional properties
  securityGroupIds?: string[];

  sentryDsn?: string;
  sentryEnv?: string;
  logLevel?: string;
}

export class GitHubStack extends cdk.Stack {
  readonly emitter: lambda.Function;
  readonly deadLetterQueue: sqs.Queue;

  constructor(scope: cdk.Construct, id: string, props: GitHubProps) {
    super(scope, id, props);
    // Validate input properties
    if (props.reportTopic === undefined && props.reportTopicARN === undefined) {
      throw Error('Either one of reportTopic and reportTopicARN must be set');
    }

    // Setup task SNS topic and SQS queue
    this.deadLetterQueue = new sqs.Queue(this, 'DeadLetterQueue')

    const taskQueueTimeout = cdk.Duration.seconds(30);
    const reportQueue = new sqs.Queue(this, 'reportQueue', {
      visibilityTimeout: taskQueueTimeout,
      deadLetterQueue: {
        maxReceiveCount: 3,
        queue: this.deadLetterQueue,
      },
    });
    const reportTopic = (props.reportTopic) ? props.reportTopic : sns.Topic.fromTopicArn(this, 'reportTopic', props.reportTopicARN!);
    reportTopic.addSubscription(new SqsSubscription(reportQueue));

    // Setup IAM role if required
    const lambdaRole = props.lambdaRoleARN ? iam.Role.fromRoleArn(this, 'LambdaRole', props.lambdaRoleARN, { mutable: false }) : undefined;

    // Setup lambda code
    const rootPath = path.resolve(__dirname, '..');
    const asset = lambda.Code.fromAsset(rootPath, {
      bundling: {
        image: lambda.Runtime.GO_1_X.bundlingDockerImage,
        user: 'root',
        command: ['go', 'build', '-o', '/asset-output/emitter', './src'],
        environment: {
          GOARCH: 'amd64',
          GOOS: 'linux',
        },
      },
    });

    /*
    if (props.vpcId) {
      vpc = ec2.Vpc.fromVpcAttributes(this, 'vpc', {
        vpcId: props.vpcId,
        privateSubnetIds: props.subnetIds,
        availabilityZones: cdk.Fn.getAzs(),
      })
    }
    */

    var securityGroups: ec2.ISecurityGroup[] | undefined = undefined;
    if (props.securityGroupIds) {
      securityGroups = props.securityGroupIds.map((sgID) => {
        return ec2.SecurityGroup.fromSecurityGroupId(this, sgID, sgID);
      });
    }

    this.emitter = new lambda.Function(this, 'emitter', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'emitter',
      code: asset,
      role: lambdaRole,
      events: [new SqsEventSource(reportQueue)],
      timeout: taskQueueTimeout,

      securityGroups: securityGroups,
      environment: {
        SECRET_ARN: props.secretARN,
        GITHUB_ENDPOINT: props.githubEndpoint || '',
        GITHUB_REPO: props.githubRepo,

        SENTRY_DSN: props.sentryDsn || "",
        SENTRY_ENVIRONMENT: props.sentryEnv || "",
        LOG_LEVEL: props.logLevel || "",
      },
    });
  }
}
