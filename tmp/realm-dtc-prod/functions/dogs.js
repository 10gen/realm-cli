exports = async (input) => {
  const { pageNumber, nPerPage } = input;
  const cluster = context.services.get('mongodb-atlas');
  const targetColl = cluster.db('verbenergy').collection('messages');
  return targetColl
    .find({ dog: true })
    .sort({ timestamp: -1 })
    .skip((pageNumber - 1) * nPerPage)
    .limit(nPerPage)
    .toArray();
};
