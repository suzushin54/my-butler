import cdk = require('@aws-cdk/core');
import { Function, Runtime, Code } from "@aws-cdk/aws-lambda"
import { RestApi, Integration, LambdaIntegration, Resource } from "@aws-cdk/aws-apigateway"
import * as iam from '@aws-cdk/aws-iam';
import * as ssm from '@aws-cdk/aws-ssm';

export class LaborServiceStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    // FIXME: You should use SecureString to store passwords, signatures and tokens..
    const channelId = ssm.StringParameter.fromStringParameterAttributes(this, 'ChannelId', {
      parameterName: '/my-butler/labor-service/channel-id',
    });
    const botId = ssm.StringParameter.fromStringParameterAttributes(this, 'BotId', {
      parameterName: '/my-butler/labor-service/bot-id',
    });
    const botOauth = ssm.StringParameter.fromStringParameterAttributes(this, 'BotOauth', {
      parameterName: '/my-butler/labor-service/bot-oauth',
    });
    const signingSecrets = ssm.StringParameter.fromStringParameterAttributes(this, 'SigningSecrets', {
      parameterName: '/my-butler/labor-service/signing-secrets',
    });
    const remoToken = ssm.StringParameter.fromStringParameterAttributes(this, 'RemoToken', {
      parameterName: '/my-butler/labor-service/remo-token',
    });

    // Create a lambda function
    const lambdaFunction: Function = new Function(this, "LaborServiceLambda", {
      functionName: "labor-service-lambda",
      runtime: Runtime.GO_1_X,
      code: Code.asset("./lambdaSource"),
      handler: "main",
      memorySize: 256,
      timeout: cdk.Duration.seconds(10),
      environment: {
        "CHANNEL_ID": channelId.stringValue,
        "BOT_ID": botId.stringValue,
        "BOT_OAUTH": botOauth.stringValue,
        "SIGNING_SECRETS": signingSecrets.stringValue,
        "REMO_TOKEN": remoToken.stringValue,
      }
    })

    // add Policy to function
    lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ["ec2:DescribeInstances"],
    }))

    // Create a API Gateway
    const restApi: RestApi = new RestApi(this, "labor-service-api", {
      restApiName: "Labor-Service-API",
      description: "Get work list from Nature Remo"
    })

    // Create a integration
    const integration: Integration = new LambdaIntegration(lambdaFunction)

    // Create a resource
    const getResource: Resource = restApi.root.addResource("event")

    // Create a POST method for slack server
    getResource.addMethod("POST", integration)
  }
}
