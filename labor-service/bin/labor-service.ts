#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { LaborServiceStack } from '../lib/labor-service-stack';

const util = require('util');
const exec = util.promisify(require('child_process').exec);

async function deploy() {
  // build golang sources
  await exec('go get -v -t -d ./lambdaSource/... && GOOS=linux GOARCH=amd64 go build -o ./lambdaSource/main ./lambdaSource/**.go');
  
  const app = new cdk.App();
  new LaborServiceStack(app, 'LaborServiceStack');
  app.synth();

  // remove a binary after build
  await exec('rm ./lambdaSource/main');
}

deploy()

