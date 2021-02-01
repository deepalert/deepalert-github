#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { GitHubStack } from '../lib/stack';

const app = new cdk.App();
new GitHubStack(app, process.env.STACK_ID!, {
    reportTopicARN: process.env.REPORT_TOPIC_ARN!,
    secretARN: process.env.SECRET_ARN!,
    githubRepo: process.env.GITHUB_REPO!,
});
