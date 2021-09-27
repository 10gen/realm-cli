exports = async () => {
  const cluster = context.services.get('mongodb-atlas');
  const targetColl = cluster.db('verbenergy').collection('messages');
  return targetColl.count({ dog: true });
};
