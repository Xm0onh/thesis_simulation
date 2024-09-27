import subprocess
import re
import os

# N_values = list(range(25, 105, 5))  # 20, 25, 30, ..., 100
N_values = [25]


F_values = {}
percentages = [0, 5, 10, 15, 20, 25, 30, 33]
for N in N_values:
    F_values[N] = [int(N * p / 100) for p in percentages]

# Refined regular expression pattern to capture TotalDuration specifically from the structured output
duration_pattern = re.compile(r"TotalDuration:([0-9.]+)s")
bandwidth_limit_pattern = re.compile(r"Bandwidth limit:  ([0-9.]+)")

# Dictionary to store the results
results = {N: {p: None for p in percentages} for N in N_values}

# Function to run the Go script with given N and F and extract TotalDuration
def run_simulation(N, F):
    try:
        result = subprocess.run(
            ["go", "run", ".", f"-N={N}", f"-f={F}"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            timeout=120  
        )

        # Find TotalDuration in the output
        duration_match = duration_pattern.search(result.stdout)
        if duration_match:
            bandwidth_limit_match = bandwidth_limit_pattern.search(result.stdout)
            if bandwidth_limit_match:
                print("Ran with N = ", N, " and F = ", F, " and got TotalDuration = ", float(duration_match.group(1)) + float(bandwidth_limit_match.group(1)))
                return float(duration_match.group(1)) + float(bandwidth_limit_match.group(1))
            else:
                return float(duration_match.group(1))
        else:
            print("TotalDuration not found in output.")
            return None
    except subprocess.TimeoutExpired:
        print(f"Timeout: N={N}, f={F}")
        return None
    except FileNotFoundError as e:
        print(f"Error: {e}. Make sure the Go executable is available in the system PATH.")
        return None
    except Exception as e:
        print(f"Unexpected error: {e}")
        return None
    finally:
        # Ensure safe termination
        subprocess.run(["pkill", "-f", "go run"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)

for N in N_values:
    for F in F_values[N]:
        if F <= N:
            print(f"Running: N={N}, F={F}")
            total_duration = run_simulation(N, F)
            
            if total_duration is not None:
                results[N][int(F * 100 / N)] = total_duration
            else:
                results[N][int(F * 100 / N)] = "Error/Timeout"

formatted_output_file = "formatted_sync_metrics_results.txt"

with open(formatted_output_file, "w") as file:
    file.write("N/F\t" + "\t".join([f"{p}%" for p in percentages]) + "\n")
    
    for N in N_values:
        row = f"{N}\t" + "\t".join([str(results[N][p]) if results[N][p] is not None else "" for p in percentages])
        file.write(row + "\n")

print("Formatted table saved in", formatted_output_file)
