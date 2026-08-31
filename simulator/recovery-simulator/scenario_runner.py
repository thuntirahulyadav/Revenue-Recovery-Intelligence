import os
import json
import requests
import pandas as pd

SIM_URL = "http://localhost:8080/api/v1/simulation/compare"

def run_benchmarking_scenario(sample_size: int = 5000):
    print(f"[Scenario Runner] Querying simulation benchmark with cohort size: {sample_size}...")
    try:
        resp = requests.get(f"{SIM_URL}?sample_size={sample_size}", timeout=10)
        if resp.status_code == 200:
            data = resp.json().get("data", {})
            baseline = data.get("baseline_strategy", {})
            ai = data.get("ai_strategy", {})
            inc = data.get("incremental_comparison", {})

            print("=" * 65)
            print(f"BENCHMARK SCENARIO REPORT ({sample_size} TRANSACTIONS)")
            print("=" * 65)
            print(f"{'Metric':<25} | {'Baseline (Blind Retry)':<18} | {'Recovery Intelligence':<18}")
            print("-" * 65)
            print(f"{'Recovery Rate':<25} | {baseline.get('recovery_rate', 0):.1%}              | {ai.get('recovery_rate', 0):.1%}")
            print(f"{'Gross Recovered (INR)':<25} | ₹{baseline.get('total_gross_recovered', 0):,.2f}     | ₹{ai.get('total_gross_recovered', 0):,.2f}")
            print(f"{'Total Action Cost (INR)':<25} | ₹{baseline.get('total_action_cost', 0):,.2f}       | ₹{ai.get('total_action_cost', 0):,.2f}")
            print(f"{'Net Recovery Value (INR)':<25} | ₹{baseline.get('net_recovery_value', 0):,.2f}     | ₹{ai.get('net_recovery_value', 0):,.2f}")
            print(f"{'Wasted Retries':<25} | {baseline.get('wasted_retries', 0):<18} | {ai.get('wasted_retries', 0):<18}")
            print("=" * 65)
            print(f"INCREMENTAL LIFT: +₹{inc.get('net_value_uplift', 0):,.2f} Net Profit ({inc.get('roi_improvement_multiple', 1)}x ROI Multiple)")
            print("=" * 65)
        else:
            print(f"[Scenario Runner] Error status {resp.status_code}: {resp.text}")
    except Exception as e:
        print(f"[Scenario Runner] Connection error: {e}")

if __name__ == "__main__":
    run_benchmarking_scenario(sample_size=5000)
