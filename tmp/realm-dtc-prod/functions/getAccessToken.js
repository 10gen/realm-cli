/* eslint-disable camelcase */
exports = async (appId) => {
  const { http } = context;
  const AUTH0_TOKEN_URL = context.values.get('AUTH0_TOKEN_URL');

  let body = { grant_type: 'client_credentials' };
  if (appId === 'EVT') {
    const client_id = context.values.get('AUTH0_EVT_CLIENT_ID');
    const client_secret = context.values.get('AUTH0_EVT_CLIENT_SECRET');
    const audience = context.values.get('AUTH0_EVT_AUDIENCE');
    if (client_id && client_secret && audience) {
      body = {
        ...body,
        client_id,
        client_secret,
        audience,
      };
    } else {
      console.error(`Requested token for app ${appId} but all required environment values are not configured`);
      return;
    }
  }

  const response = await http.post({
    url: AUTH0_TOKEN_URL,
    body,
    encodeBodyAsJSON: true,
  });
  if (response.statusCode !== 200) {
    console.error('Access token request failed', response);
    return;
  }
  const responseBody = EJSON.parse(response.body.text());
  // eslint-disable-next-line consistent-return
  return responseBody.access_token;
};
