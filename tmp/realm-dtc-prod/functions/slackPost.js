// Send a message to a Slack Channel
// Example usage:
//   slackPost({ text: 'Hello World', channel: '<channel-id>' })
exports = async (payload) => {
  const SLACK_API_BASE_URL = context.values.get('SLACK_API_BASE_URL');
  const SLACK_API_TOKEN = context.values.get('SLACK_API_TOKEN');
  const { http } = context;
  const body = {
    ...payload,
    as_user: !!payload.as_user,
    token: payload.token || SLACK_API_TOKEN,
    username: payload.username || 'Verb Bot',
  };
  const headers = { 'Content-Type': ['application/json'], Authorization: [`Bearer ${SLACK_API_TOKEN}`] };
  const response = await http.post({
    body,
    encodeBodyAsJSON: true,
    headers,
    url: `${SLACK_API_BASE_URL}/chat.postMessage`,
  });
  const slackResponseData = EJSON.parse(response.body.text());
  const _stringResponse = JSON.stringify(slackResponseData);
  if (!slackResponseData.ok) {
    console.error(_stringResponse);
    throw Error(`Error Sending Slack Message --- ${slackResponseData.error}`);
  }
  console.log(_stringResponse);
  return true;
};
