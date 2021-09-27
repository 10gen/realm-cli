/*
  This function constructs a cart for a given order request.

  TODO: Make a custom resolver for this.

  returns: a subset of a newOrder object representing the order cart
*/

exports = async (input) => {
  const { customer, order } = input;
  try {
    const GOOGLE_ANALYTICS_ID = context.values.get('GOOGLE_ANALYTICS_ID');
    const querystring = require('querystring');
    const { http } = context;

    const data = {
      v: '1',
      tid: GOOGLE_ANALYTICS_ID,
      cid: customer._id.toString(),
      uid: customer._id.toString(),
      cd1: customer._id.toString(),
      t: 'event',
      ec: 'Ecommerce MP',
      ea: 'Completed Transaction',
      ti: order.invoiceNumber,
      ta: `Verb ${order.orderType} MP`,
      tr: context.functions.execute('centsToDollars', order.paid.total),
      tt: 0,
      ts: context.functions.execute('centsToDollars', order.paid.shipping),
      dh: 'crm.verbenergy.co',
      pa: 'purchase',
    };
    order.items.forEach((item, idx) => {
      const pr = `pr${idx + 1}`;
      data[`${pr}id`] = item.id;
      data[`${pr}nm`] = item.name;
      data[`${pr}ca`] = 'Caffeinated Energy Bars';
      data[`${pr}pr`] = context.functions.execute('centsToDollars', item.price);
      data[`${pr}qt`] = item.quantity;
    });

    const body = querystring.stringify(data);
    const response = await http.post({
      url: 'https://www.google-analytics.com/collect',
      body,
    });
    if (response.statusCode !== 200) {
      throw new Error(response.body.text());
    }
  } catch (e) {
    console.error(`Google analytics collection failed for ${order.invoiceNumber}`, e);
  }
};
