// Get list of Slack Channels
exports = async () => {
  const SLACK_API_BASE_URL = context.values.get('SLACK_API_BASE_URL');
  const SLACK_API_TOKEN = context.values.get('SLACK_API_TOKEN');
  const { http } = context;
  const headers = { 'Content-Type': ['application/json'], Authorization: [`Bearer ${SLACK_API_TOKEN}`] };
  const response = await http.get({
    headers,
    url: `${SLACK_API_BASE_URL}/conversations.list?limit=1000&exclude_archived=true`,
  });
  const slackResponseData = EJSON.parse(response.body.text());
  const _stringResponse = JSON.stringify(slackResponseData);
  if (!slackResponseData.ok) {
    console.error(_stringResponse);
    throw Error(`Error Retrieving Channels --- ${slackResponseData.error}`);
  }
  return slackResponseData.channels;
};
