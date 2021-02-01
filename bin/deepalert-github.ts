#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { GitHubStack } from '../lib/stack';

const app = new cdk.App();
new GitHubStack(app, process.env.STACK_ID!, {
    reportTopicARN: 'arn:aws:sns:us-east-1:111122223333:my-topic',
    secretARN: 'test-secret-arn',
    githubRepo: 'test/repository',
    securityGroupIds: ['sg-1'],
});
