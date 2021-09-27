/*
  This function converts cents to dollars
*/

exports = (valueCents) => {
  if (!valueCents) {
    return '';
  }
  return (valueCents / 100).toFixed(2);
};
