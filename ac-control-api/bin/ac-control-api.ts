#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { AcControlApiStack } from '../lib/ac-control-api-stack';

const util = require('util');
const exec = util.promisify(require('child_process').exec);

async function deploy() {
    // build go sources
    await exec('go get -v -t -d ./lambdaSource/... && GOOS=linux GOARCH=amd64 go build -o ./lambdaSource/main ./lambdaSource/**.go');

    const app = new cdk.App();
    new AcControlApiStack(app, 'AcControlApiSlack');
    app.synth();

    // remove binary of build results
    await exec('rm ./lambdaSource/main');
}

deploy()