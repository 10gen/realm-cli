exports = (collectionName, { limit = 100, search, skip = 0 }) => {
  const collection = context.services.get('mongodb-atlas').db(context.values.get('DB_NAME')).collection(collectionName);
  return collection.find(
    { $text: { $search: search } },
    { score: { $meta: 'textScore' } },
  ).sort({ score: { $meta: 'textScore' } })
    .skip(skip)
    .limit(limit)
    .toArray();
};
