#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { LaborServiceStack } from '../lib/labor-service-stack';

const app = new cdk.App();
new LaborServiceStack(app, 'LaborServiceStack');
