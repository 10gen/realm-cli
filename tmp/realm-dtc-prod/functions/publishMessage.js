exports = async (message, routingKey) => {
  const { CLOUDAMQP_URL, CLOUDAMQP_USER } = context.environment.values;
  const CLOUDAMQP_PW = context.values.get('CLOUDAMQP_PW');
  const EXCHANGE = context.values.get('VERB_DIRECT_EXCHANGE');
  const { http } = context;

  const body = {
    properties: { content_type: 'application/json' },
    routing_key: routingKey,
    payload: JSON.stringify(message),
    payload_encoding: 'string',
  };

  const basicCreds = Buffer.from(`${CLOUDAMQP_USER}:${CLOUDAMQP_PW}`).toString('base64');
  const headers = {
    Authorization: [`Basic ${basicCreds}`],
  };
  const response = await http.post({
    url: `${CLOUDAMQP_URL}/api/exchanges/${CLOUDAMQP_USER}/${EXCHANGE}/publish`,
    headers,
    body,
    encodeBodyAsJSON: true,
  });

  return response.statusCode === 200;
};
