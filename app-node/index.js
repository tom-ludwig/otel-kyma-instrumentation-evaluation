"use strict";

const express = require("express");
const { Pool } = require("pg");

const app = express();
const port = 8080;

const pool = new Pool({
  host: process.env.PG_HOST || "postgres",
  port: parseInt(process.env.PG_PORT || "5432", 10),
  database: process.env.PG_DB || "postgres",
  user: process.env.PG_USER || "postgres",
  password: process.env.PG_PASSWORD || "postgres",
});

pool.on("error", (err) => {
  console.error("idle pg client error:", err.message);
});

app.get("/rolldice", async (req, res) => {
  const roll = Math.floor(Math.random() * 6) + 1;

  // Simulate some latency, mirroring the Go app
  await new Promise((resolve) =>
    setTimeout(resolve, Math.floor(Math.random() * 50)),
  );

  // Issue a trivial DB query so the trace has a child DB span.
  try {
    await pool.query("SELECT 1");
  } catch (err) {
    console.error("db query failed:", err.message);
    res.status(500).send(`db error: ${err.message}\n`);
    return;
  }

  res.send(`rolled: ${roll}\n`);
});

app.get("/healthz", (req, res) => {
  res.sendStatus(200);
});

app.listen(port, () => {
  console.log(`listening on :${port}`);
});
