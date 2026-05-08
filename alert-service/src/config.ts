export interface Config {
  port: number;
  logLevel: string;
  apiGatewayUrl: string;
}

export function loadConfig(): Config {
  return {
    port: parseInt(process.env.ALERT_PORT || '8002', 10),
    logLevel: process.env.LOG_LEVEL || 'INFO',
    apiGatewayUrl: process.env.API_GATEWAY_URL || 'http://api-gateway:8000',
  };
}
