exports = (query) => {
  if ('activeUser_exists' in query) {
    // Convert realm's generated QueryInput fields to mongodb query params
    // This is a workaround to a problem that is better articulated in this thread:
    // https://developer.mongodb.com/community/forums/t/custom-query-resolvers-supporting-the-generated-query-input/14274
    query.activeUser = { $exists: query.activeUser_exists };
    delete query.activeUser_exists;
  }
  if ('updated_gte' in query) {
    // Convert realm's generated QueryInput fields to mongodb query params
    // This is a workaround to a problem that is better articulated in this thread:
    // https://developer.mongodb.com/community/forums/t/custom-query-resolvers-supporting-the-generated-query-input/14274
    query.updated = { $gte: query.updated_gte };
    delete query.updated_gte;
  }
  const collection = context.services.get('mongodb-atlas').db('verbenergy').collection('tickets');
  return collection.aggregate([
    { $match: query },
    { $group: { _id: '$type', count: { $sum: 1 }, type: { $first: '$type' } } },
    { $sort: { _id: 1 } },
  ]).toArray();
};
