// Sending a general Text Message to specified slack channel.
exports = async ({
  messageId, channelId, hasImage, comments,
}) => {
  const _id = new BSON.ObjectId(messageId);
  const VERB_CRM_BASE_URL = context.values.get('VERB_CRM_BASE_URL');
  const collection = context.services.get('mongodb-atlas').db('verbenergy').collection('messages');
  const message = await collection.findOne({ _id });
  const { body, mediaURL, customer } = message;
  const payload = { channel: channelId };

  if (hasImage) {
    payload.attachments = [{ fallback: 'Image', image_url: mediaURL }];
    payload.text = comments || '';
  } else {
    payload.text = `${body}\n--${comments ? `\n${comments}` : ''}`;
  }

  // add the link to the customer's page
  payload.text += `\n customer link: ${VERB_CRM_BASE_URL}/?phone=${customer}`;
  await context.functions.execute('slackPost', payload);

  return message;
};
