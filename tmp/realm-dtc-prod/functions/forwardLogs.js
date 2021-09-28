exports = async (logs) => {
  const DD_LOG_URL = context.values.get('DD_LOG_URL');
  const DD_API_KEY = context.values.get('DD_API_KEY');
  const logColl = context.services.get('mongodb-atlas')
    .db(context.values.get('DB_NAME'))
    .collection('realmlogs');

  const body = logs.map((log) => ({
    ddsource: 'realm',
    ddtags: `env:${context.environment.tag}`,
    message: JSON.stringify(log),
    service: 'realm-dtc',
  }));

  const response = await context.http.post({
    url: DD_LOG_URL,
    headers: {
      'Content-Type': ['application/json'],
      'DD-API-KEY': [DD_API_KEY],
    },
    body,
    encodeBodyAsJSON: true,
  });
  if (response.statusCode !== 200) {
    console.error('Error forwarding logs to Datadog.');
  } else {
    await Promise.all(logs.map(async (log) => {
      await logColl.updateOne({ _id: log._id }, { $set: { forwarded_to_dd: true } });
    }));
  }
};
