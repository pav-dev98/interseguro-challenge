const assert = require("node:assert/strict");
const test = require("node:test");
const { calculateStatistics, isDiagonal } = require("../dist/statistics/statistics");

test("calculates max, min, sum and average across Q and R", () => {
  const statistics = calculateStatistics(
    [[0.5, 1.5], [-1, 2]],
    [[3, 0], [0, 4]]
  );

  assert.equal(statistics.max, 4);
  assert.equal(statistics.min, -1);
  assert.equal(statistics.sum, 10);
  assert.equal(statistics.average, 1.25);
  assert.equal(statistics.hasDiagonalMatrix, true);
});

test("recognizes a diagonal matrix with off-diagonal values within tolerance", () => {
  assert.equal(isDiagonal([[2, 1e-13], [-1e-13, 3]]), true);
});

test("does not consider non-square or non-diagonal matrices as diagonal", () => {
  assert.equal(isDiagonal([[1, 0, 0], [0, 1, 0]]), false);
  assert.equal(isDiagonal([[1, 0.1], [0, 1]]), false);
  assert.equal(calculateStatistics([[1, 2], [3, 4]], [[1, 2], [3, 4]]).hasDiagonalMatrix, false);
});

test("rejects empty, irregular and non-finite matrices", () => {
  assert.throws(() => calculateStatistics([], [[1]]), /q must be a non-empty matrix/);
  assert.throws(() => calculateStatistics([[1, 2], [3]], [[1]]), /q rows must have the same length/);
  assert.throws(() => calculateStatistics([[Number.NaN]], [[1]]), /finite numbers/);
  assert.throws(() => calculateStatistics([[Infinity]], [[1]]), /finite numbers/);
});
