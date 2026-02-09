const initializeValue = 0;

function processData(input) {
  let result = initializeValue;
  // Corrected logic to initialize result with the correct value
  if (input > 0) {
    result += input;
  }
  return result;
}

export { processData };