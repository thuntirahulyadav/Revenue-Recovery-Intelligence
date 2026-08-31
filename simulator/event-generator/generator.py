import time
import random
import requests
import uuid
import sys

API_URL = "http://localhost:8080/api/v1/events/payment-failed"

FAILURE_REASONS = [
    "BANK_TIMEOUT",
    "NETWORK_ERROR",
    "INSUFFICIENT_FUNDS",
    "CARD_EXPIRED",
    "PAYMENT_METHOD_FAILURE",
    "CUSTOMER_ABANDONMENT",
    "TECHNICAL_ERROR",
]
WEIGHTS = [0.25, 0.20, 0.20, 0.10, 0.12, 0.08, 0.05]

METHODS = ["card", "upi", "netbanking", "wallet", "emi"]

def emit_continuous_traffic(interval_seconds: float = 1.5, total_events: int = 50):
    print(f"[Simulator Traffic] Starting continuous failure stream to {API_URL} (interval: {interval_seconds}s)")
    count = 0
    while count < total_events:
        count += 1
        amt = random.choices([
            random.randint(400, 3500),
            random.randint(5000, 25000),
            random.randint(30000, 85000),
        ], weights=[0.60, 0.32, 0.08])[0]

        reason = random.choices(FAILURE_REASONS, weights=WEIGHTS)[0]
        method = random.choice(METHODS)
        attempts = random.choices([1, 2, 3], weights=[0.7, 0.2, 0.1])[0]

        payload = {
            "amount": amt,
            "payment_method": method,
            "failure_reason": reason,
            "attempt_count": attempts,
        }

        try:
            resp = requests.post(API_URL, json=payload, timeout=4)
            if resp.status_code in [200, 201]:
                data = resp.json().get("data", {})
                dec = data.get("decision", {})
                pred = data.get("prediction", {})
                print(f"[{count}/{total_events}] Emitted: ₹{amt} ({reason}) -> Strategy: {dec.get('strategy')} | Prob: {pred.get('recovery_probability', 0):.0%} | Net: ₹{dec.get('expected_net_value', 0):.2f}")
            else:
                print(f"[{count}] Non-200 response: {resp.status_code}")
        except Exception as e:
            print(f"[{count}] Emission note: {e}")

        time.sleep(interval_seconds)

if __name__ == "__main__":
    count = int(sys.argv[1]) if len(sys.argv) > 1 else 20
    emit_continuous_traffic(interval_seconds=1.0, total_events=count)
