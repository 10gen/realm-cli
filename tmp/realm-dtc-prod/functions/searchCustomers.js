exports = (searchInput) => {
  console.log('Searching Customers with these fields: ', JSON.stringify(searchInput));
  return context.functions.execute('searchCollection', 'newcustomers', searchInput);
};
