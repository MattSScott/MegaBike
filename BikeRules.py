import os

ITERS = 30
SCAR = [0.5, 1.0, 1.5, 2.0, 2.5]
MUT = "IMMUTABLE"


def run(scarcity):
    script = f"go run . -agents=100 -loot={scarcity}"
    avgLife = 0.0
    for _ in range(ITERS):
        os.system(script)
        with open("output.txt", "r") as FILE:
            life = FILE.read()
            avgLife += float(life)
    avgLife /= ITERS
    return avgLife


def main():
    results = {}
    for scar in SCAR:
        ans = run(scar)
        results[(MUT, scar)] = ans
    print(results)


main()
