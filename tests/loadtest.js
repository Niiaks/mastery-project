import http from "k6/http";

import { check, sleep } from "k6";

export let options = {
  stages: [
    { duration: "30s", target: 10 }, // ramp up
    { duration: "30s", target: 20 },
    { duration: "30s", target: 30 },
    { duration: "30s", target: 0 }, // ramp down
  ],
};

export default function () {
  let res = http.get("https://itemapi.tech/api/v1/health");
  check(res, {
    "status is 200 or 429": (r) => r.status === 200 || r.status === 429,
  });
  sleep(10);
}
