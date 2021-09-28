/*
  This utility function maps customer query to properties formatted for event emission
*/

exports = (customerQuery) => {
  const properties = {};
  Object.keys(customerQuery).forEach((key) => {
    properties[`query_${key}`] = customerQuery[key];
  });
  return properties;
};
