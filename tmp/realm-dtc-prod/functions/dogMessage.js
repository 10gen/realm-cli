// Set or Unset Dog Field on a Text Message.
exports = async ({ messageId, dog = true }) => {
  const _id = new BSON.ObjectId(messageId);
  const collection = context.services.get('mongodb-atlas').db('verbenergy').collection('messages');
  const message = await collection.findOneAndUpdate(
    { _id },
    { $set: { dog } },
    { returnNewDocument: true },
  );
  console.log(JSON.stringify(message));
  return message;
};
