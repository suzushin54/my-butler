#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { AcControlApiStack } from '../lib/ac-control-api-stack';

const app = new cdk.App();
new AcControlApiStack(app, 'AcControlApiStack');
