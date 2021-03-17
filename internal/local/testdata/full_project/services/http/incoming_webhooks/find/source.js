
exports = function({ query }) {
    const {a, b, c} = query

    const filter = {}
    if (!!a) {
      filter.a = a
    }
    if (!!b) {
      filter.b = b
    }
    if (!!c) {
      filter.c = c
    }

    return context.services
      .get('mongodb-atlas')
      .db('test')
      .collection('coll2')
      .find(filter)
};
