const assert = require("node:assert/strict");
const test = require("node:test");
const { app } = require("../dist/app");

test("GET /internal/health returns the Node API status", async () => {
  const server = app.listen(0);
  const { port } = server.address();

  try {
    const response = await fetch(`http://127.0.0.1:${port}/internal/health`);

    assert.equal(response.status, 200);
    assert.deepEqual(await response.json(), {
      status: "active",
      service: "node-api"
    });
  } finally {
    await new Promise((resolve) => server.close(resolve));
  }
});
