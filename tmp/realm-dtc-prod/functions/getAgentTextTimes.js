/*
*/

/*
 * Requires the MongoDB Node.js Driver
 * https://mongodb.github.io/node-mongodb-native
 */

exports = async ({
  forDate = undefined,
  // if date supplied, default date range is 1
  days = forDate ? 1 : 7,
  offset = 0,
} = {}) => {
  const moment = require('moment');

  const today = moment(forDate).endOf('day');
  const maxDate = today.clone().subtract(offset * days, 'days');
  const minDate = maxDate.clone().subtract(days, 'days');

  const db = context.services.get('mongodb-atlas').db('verbenergy');
  const coll = db.collection('users');

  const agg = [
    {
      // Reduce document size to essential fields
      $project: {
        email: 1,
        name: 1,
      },
    }, {
      // join messages to users, grouped by hour and date
      $lookup: {
        from: 'messages',
        let: { id: '$_id' }, // user._id
        pipeline: [
          {
            $match: {
              $expr: { $eq: ['$user', '$$id'] },
              timestamp: {
                $gt: minDate.toDate(),
                $lt: maxDate.toDate(),
              },
            },
          }, {
            // remove extra attrs
            $project: {
              timestamp: 1,
              customerId: 1,
            },
          }, {
            // allow subsequent groups to be in chronological order
            $sort: { timestamp: 1 },
          }, {
            $facet: {
              messageCount: [
                { $count: 'count' },
              ],
              // histogram view
              groupByHour: [
                {
                  $group: {
                    _id: {
                      $dateFromParts: {
                        year: { $year: '$timestamp' },
                        month: { $month: '$timestamp' },
                        day: { $dayOfMonth: '$timestamp' },
                        hour: { $hour: '$timestamp' },
                      },
                    },
                    count: { $sum: 1 },
                    earliest: { $min: '$timestamp' },
                    latest: { $max: '$timestamp' },
                  },
                }, {
                  $sort: { _id: 1 }, // date-hour
                },
              ],
              // day view
              dayStats: [
                {
                  $group: {
                    _id: {
                      $dateFromParts: {
                        year: { $year: '$timestamp' },
                        month: { $month: '$timestamp' },
                        day: { $dayOfMonth: '$timestamp' },
                      },
                    },
                    count: { $sum: 1 },
                    earliest: { $min: '$timestamp' },
                    latest: { $max: '$timestamp' },
                  },
                },
              ],
            },
          },
        ],
        as: 'messages',
      },
    }, {
      // hoist nested $messages object to root document
      $unwind: { path: '$messages' },
    }, {
      $addFields: {
        activeHours: { $size: '$messages.groupByHour' },
        activeDays: { $size: '$messages.dayStats' },
        sentMessages: { $first: '$messages.messageCount.count' },
      },
    }, {
      // remove users with no activity within period
      $match: { sentMessages: { $gt: 0 } },
    }, {
      $project: {
        email: 1,
        name: 1,
        activeHours: 1,
        activeDays: 1,
        sentMessages: 1,
        hours: '$messages.groupByHour',
        days: '$messages.dayStats',
      },
    },
  ];
  let results = [];

  try {
    results = (await coll.aggregate(agg)).toArray();
  } catch (error) {
    console.log(error.message);
  }

  return results;
};
