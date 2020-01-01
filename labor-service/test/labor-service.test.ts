import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import cdk = require('@aws-cdk/core');
import LaborService = require('../lib/labor-service-stack');

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new LaborService.LaborServiceStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});