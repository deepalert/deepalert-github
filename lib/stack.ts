import * as cdk from '@aws-cdk/core';
import * as lambda from '@aws-cdk/aws-lambda';
import * as iam from '@aws-cdk/aws-iam';
import * as sns from '@aws-cdk/aws-sns';
import * as sqs from '@aws-cdk/aws-sqs';
import { SqsEventSource } from '@aws-cdk/aws-lambda-event-sources';
import { SqsSubscription } from '@aws-cdk/aws-sns-subscriptions';

import * as path from 'path';

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

    this.emitter = new lambda.Function(this, 'emitter', {
      runtime: lambda.Runtime.GO_1_X,
      handler: 'emitter',
      code: asset,
      role: lambdaRole,
      events: [new SqsEventSource(reportQueue)],
      timeout: taskQueueTimeout,
      environment: {
        SECRET_ARN: props.secretARN,
        GITHUB_ENDPOINT: props.githubEndpoint || 'https://api.github.com',
        GITHUB_REPO: props.githubRepo,

        SENTRY_DSN: props.sentryDsn || "",
        SENTRY_ENVIRONMENT: props.sentryEnv || "",
        LOG_LEVEL: props.logLevel || "",
      },
    });
  }
}
