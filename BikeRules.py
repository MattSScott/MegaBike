import os
import matplotlib.pyplot as plt
import seaborn

ITERS = 10
SCAR = [0.0, 0.5, 1.0, 1.5, 2.0, 2.5]
MUT = "MUTABLE"


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


def gen_results():
    results = {}
    for scar in SCAR:
        ans = run(scar)
        results[(MUT, scar)] = ans
    print(results)


def main():
    # mutable = {
    #     ("MUTABLE", 0.0): 22.790606666666665,
    #     ("MUTABLE", 0.5): 47.76049999999999,
    #     ("MUTABLE", 1.0): 71.28070666666667,
    #     ("MUTABLE", 1.5): 83.35808666666667,
    #     ("MUTABLE", 2.0): 88.76201000000002,
    #     ("MUTABLE", 2.5): 91.17147666666665,
    # }

    mutable = {
        ("MUTABLE", 0.0): 22.91135,
        ("MUTABLE", 0.5): 44.342740000000006,
        ("MUTABLE", 1.0): 63.997809999999994,
        ("MUTABLE", 1.5): 73.15179,
        ("MUTABLE", 2.0): 78.00287,
        ("MUTABLE", 2.5): 80.30702000000001,
    }

    # immutable = {
    #     ("IMMUTABLE", 0.0): 22.74136666666667,
    #     ("IMMUTABLE", 0.5): 20.685543333333335,
    #     ("IMMUTABLE", 1.0): 21.04168666666667,
    #     ("IMMUTABLE", 1.5): 21.336950000000005,
    #     ("IMMUTABLE", 2.0): 21.927050000000005,
    #     ("IMMUTABLE", 2.5): 22.52041666666667,
    # }
    immutable = {
        ("IMMUTABLE", 0.0): 22.886969999999998,
        ("IMMUTABLE", 0.5): 21.73035,
        ("IMMUTABLE", 1.0): 25.102809999999998,
        ("IMMUTABLE", 1.5): 29.729470000000003,
        ("IMMUTABLE", 2.0): 34.13888,
        ("IMMUTABLE", 2.5): 43.085950000000004,
    }

    data = [list(mutable.values()), list(immutable.values())]
    x_lab = [0, 0.5, 1.0, 1.5, 2.0, 2.5]
    y_lab = ["Mutable", "Immutable"]

    seaborn.heatmap(data, xticklabels=x_lab, yticklabels=y_lab, square=True)
    plt.xlabel("Ratio of Resouces to Agents")
    plt.ylabel("Rule Mutability")
    plt.title("Average Number of Iterations Survived")
    plt.show()


main()
# gen_results()
