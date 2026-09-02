const assert = require("node:assert/strict");
const test = require("node:test");
const { app } = require("../dist/app");

async function withServer(callback) {
  const server = app.listen(0);
  const { port } = server.address();
  try {
    await callback(`http://127.0.0.1:${port}`);
  } finally {
    await new Promise((resolve) => server.close(resolve));
  }
}

async function post(baseURL, body) {
  return fetch(`${baseURL}/statistics`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body
  });
}

test("POST /statistics returns calculated statistics for valid matrices", async () => {
  await withServer(async (baseURL) => {
    const response = await post(baseURL, JSON.stringify({ q: [[1, 0], [0, 1]], r: [[2, 3], [0, 4]] }));

    assert.equal(response.status, 200);
    assert.deepEqual(await response.json(), {
      max: 4,
      min: 0,
      sum: 11,
      average: 11 / 8,
      hasDiagonalMatrix: true
    });
  });
});

test("POST /statistics rejects invalid input", async (t) => {
  const cases = [
    { name: "invalid JSON", body: "{\"q\":" },
    { name: "missing q", body: JSON.stringify({ r: [[1]] }) },
    { name: "missing r", body: JSON.stringify({ q: [[1]] }) },
    { name: "invalid matrices", body: JSON.stringify({ q: [[1, 2], [3]], r: [[1]] }) }
  ];

  for (const item of cases) {
    await t.test(item.name, async () => {
      await withServer(async (baseURL) => {
        const response = await post(baseURL, item.body);
        assert.equal(response.status, 400);
        assert.equal(typeof (await response.json()).error, "string");
      });
    });
  }
});
