import cdk = require('@aws-cdk/core');
import * as iam from '@aws-cdk/aws-iam';
import * as ssm from '@aws-cdk/aws-ssm';
import { Function, Runtime, Code} from "@aws-cdk/aws-lambda"
import { RestApi, Integration, LambdaIntegration, Resource } from "@aws-cdk/aws-apigateway"

export class AcControlApiStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const signingSecrets = ssm.StringParameter.fromStringParameterAttributes(this, 'SigningSecrets', {
      parameterName: '/my-butler/labor-service/signing-secrets',
    });
    const remoToken = ssm.StringParameter.fromStringParameterAttributes(this, 'RemoToken', {
      parameterName: '/my-butler/labor-service/remo-token',
    });
    const applianceId = ssm.StringParameter.fromStringParameterAttributes(this, 'ApplianceId', {
      parameterName: '/my-butler/labor-service/appliance-id',
    });

    const iftttApiKey = ssm.StringParameter.fromStringParameterAttributes(this, 'IftttApiKey', {
      parameterName: '/my-butler/labor-service/ifttt-api-key',
    });

    // Create lambda function
    const lambdaFunction: Function = new Function(this, "ac-control-function", {
      functionName: "ac-control-function",
      runtime: Runtime.GO_1_X,
      code: Code.asset("./lambdaSource"),
      handler: "main",
      memorySize: 256,
      timeout: cdk.Duration.seconds(10),
      environment: {
        "SIGNING_SECRETS": signingSecrets.stringValue,
        "REMO_TOKEN": remoToken.stringValue,
        "APPLIANCE_ID": applianceId.stringValue,
        "IFTTT_API_KEY": iftttApiKey.stringValue,
      }
    })

    // Add policy to function
    lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ["ec2:StartInstances", "ec2:StopInstances", "ec2:DescribeInstance"],
    }))

    // Create API Gateway
    const restApi: RestApi = new RestApi(this, "ac-control-api", {
      restApiName: "ac-control-api",
      description: "Home AC control via nature remo"
    })

    // Create integration
    const integration: Integration = new LambdaIntegration(lambdaFunction)

    // Create resource
    const getResource: Resource = restApi.root.addResource("event")

    // Create method
    getResource.addMethod("POST", integration)

  }
}
