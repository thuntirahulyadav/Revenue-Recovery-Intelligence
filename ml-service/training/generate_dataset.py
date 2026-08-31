import os
import uuid
import random
import numpy as np
import pandas as pd
from typing import Optional

RANDOM_SEED = 42

PAYMENT_METHODS = ["card", "upi", "netbanking", "wallet", "emi"]
PAYMENT_METHOD_WEIGHTS = [0.45, 0.35, 0.10, 0.06, 0.04]

FAILURE_REASONS = [
    "BANK_TIMEOUT",
    "NETWORK_ERROR",
    "INSUFFICIENT_FUNDS",
    "CARD_EXPIRED",
    "PAYMENT_METHOD_FAILURE",
    "CUSTOMER_ABANDONMENT",
    "TECHNICAL_ERROR",
]
FAILURE_REASON_WEIGHTS = [0.22, 0.18, 0.20, 0.08, 0.14, 0.10, 0.08]

STRATEGIES = [
    "RETRY_NOW",
    "RETRY_LATER",
    "SWITCH_PAYMENT_METHOD",
    "SEND_PAYMENT_LINK",
    "ESCALATE_TO_HUMAN",
    "STOP_RECOVERY",
]

def generate_synthetic_dataset(num_records: int = 12000, seed: int = RANDOM_SEED) -> pd.DataFrame:
    np.random.seed(seed)
    random.seed(seed)

    # 1. Create a pool of merchants and customers
    num_merchants = 25
    merchant_ids = [str(uuid.UUID(int=i + 1)) for i in range(num_merchants)]
    
    num_customers = 1500
    customer_pool = []
    for i in range(num_customers):
        c_id = str(uuid.UUID(int=1000 + i))
        m_id = random.choice(merchant_ids)
        # Bimodal customer value (standard vs VIP)
        if random.random() < 0.15:
            c_val = np.random.uniform(50000, 300000)
            c_succ = np.random.beta(8, 2)  # Skewed high
        else:
            c_val = np.random.uniform(500, 30000)
            c_succ = np.random.beta(5, 2)
        
        c_succ = float(np.clip(c_succ, 0.1, 0.99))
        c_fail = float(np.clip(1.0 - c_succ, 0.01, 0.9))
        customer_pool.append({
            "customer_id": c_id,
            "merchant_id": m_id,
            "customer_success_rate": round(c_succ, 4),
            "customer_failure_rate": round(c_fail, 4),
            "customer_value": round(c_val, 2),
        })

    records = []

    for i in range(num_records):
        payment_id = str(uuid.UUID(int=100000 + i))
        cust = random.choice(customer_pool)
        
        # Payment amount (log-normal distribution for realistic transaction values)
        # Amounts range typically from ₹100 to ₹100,000+ (values stored in INR units)
        amt_raw = np.random.lognormal(mean=7.5, sigma=1.2) # median ~ ₹1800
        amount = int(np.clip(amt_raw, 100, 150000))

        payment_method = random.choices(PAYMENT_METHODS, weights=PAYMENT_METHOD_WEIGHTS)[0]
        failure_reason = random.choices(FAILURE_REASONS, weights=FAILURE_REASON_WEIGHTS)[0]
        
        # Attempt count
        attempt_count = random.choices([1, 2, 3, 4, 5], weights=[0.60, 0.22, 0.11, 0.05, 0.02])[0]

        hour_weights = np.array([
            0.01, 0.01, 0.005, 0.005, 0.005, 0.01, 0.02, 0.03,
            0.05, 0.07, 0.08, 0.09, 0.08, 0.07, 0.06, 0.06,
            0.07, 0.08, 0.09, 0.08, 0.06, 0.04, 0.02, 0.015
        ])
        hour_probs = hour_weights / hour_weights.sum()
        hour_of_day = int(np.random.choice(range(24), p=hour_probs))
        day_of_week = random.randint(0, 6) # 0=Monday, 6=Sunday

        # Strategy selection simulation
        # Determine candidate strategy and recovery probability
        base_recovery_p = 0.0

        if failure_reason == "BANK_TIMEOUT":
            strategy = "RETRY_LATER" if attempt_count <= 2 else "SWITCH_PAYMENT_METHOD"
            base_recovery_p = 0.72 - (attempt_count * 0.10)
        elif failure_reason == "NETWORK_ERROR":
            strategy = "RETRY_NOW" if attempt_count == 1 else "RETRY_LATER"
            base_recovery_p = 0.82 - (attempt_count * 0.12)
        elif failure_reason == "INSUFFICIENT_FUNDS":
            strategy = "RETRY_LATER"
            base_recovery_p = 0.46 - (attempt_count * 0.08)
            # Higher chance on 1st/30th of month or evening
            if hour_of_day >= 18:
                base_recovery_p += 0.08
        elif failure_reason == "CARD_EXPIRED":
            strategy = "SWITCH_PAYMENT_METHOD" if random.random() < 0.5 else "SEND_PAYMENT_LINK"
            base_recovery_p = 0.65
        elif failure_reason == "CUSTOMER_ABANDONMENT":
            strategy = "SEND_PAYMENT_LINK"
            base_recovery_p = 0.60
        elif failure_reason == "PAYMENT_METHOD_FAILURE":
            strategy = "SWITCH_PAYMENT_METHOD"
            base_recovery_p = 0.68 - (attempt_count * 0.07)
        elif failure_reason == "TECHNICAL_ERROR":
            strategy = "RETRY_NOW" if attempt_count == 1 else "RETRY_LATER"
            base_recovery_p = 0.70 - (attempt_count * 0.10)
        else:
            strategy = "STOP_RECOVERY"
            base_recovery_p = 0.10

        # High value override
        if amount > 75000:
            strategy = "ESCALATE_TO_HUMAN"
            base_recovery_p += 0.10

        # Factor in customer historical reliability
        p_final = base_recovery_p * 0.65 + (cust["customer_success_rate"] * 0.35)
        
        # Penalize repeated attempts
        if attempt_count >= 4:
            p_final *= 0.4
            if random.random() < 0.4:
                strategy = "STOP_RECOVERY"

        p_final = float(np.clip(p_final, 0.02, 0.96))

        # Recovery attempted & success
        recovery_attempted = strategy != "STOP_RECOVERY"
        if recovery_attempted:
            recovery_success = 1 if (random.random() < p_final) else 0
        else:
            recovery_success = 0

        records.append({
            "payment_id": payment_id,
            "merchant_id": cust["merchant_id"],
            "customer_id": cust["customer_id"],
            "transaction_amount": amount,
            "payment_method": payment_method,
            "failure_reason": failure_reason,
            "attempt_count": attempt_count,
            "customer_success_rate": cust["customer_success_rate"],
            "customer_failure_rate": cust["customer_failure_rate"],
            "customer_value": cust["customer_value"],
            "hour_of_day": hour_of_day,
            "day_of_week": day_of_week,
            "recovery_attempted": recovery_attempted,
            "recovery_strategy": strategy,
            "recovery_success": recovery_success,
        })

    df = pd.DataFrame(records)
    return df

def save_synthetic_dataset(output_path: Optional[str] = None):
    if output_path is None:
        base_dir = os.path.dirname(os.path.abspath(__file__))
        output_dir = os.path.join(base_dir, "..", "data", "synthetic")
        os.makedirs(output_dir, exist_ok=True)
        output_path = os.path.join(output_dir, "payments_recovery_synthetic.csv")
    
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    df = generate_synthetic_dataset(num_records=12500, seed=RANDOM_SEED)
    df.to_csv(output_path, index=False)
    print(f"[Dataset Generator] Successfully generated {len(df)} records at: {output_path}")
    print(f"[Dataset Generator] Recovery Success Rate: {df['recovery_success'].mean():.2%}")
    print("[Dataset Generator] Failure distribution:\n", df['failure_reason'].value_counts())
    return output_path

if __name__ == "__main__":
    save_synthetic_dataset()
