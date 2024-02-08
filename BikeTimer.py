import os
import time
import matplotlib.pyplot as plt
import numpy as np


def run(n_agents=50, n_rules=10, is_strat="true"):

    full_cmd = f"go run . -agents={n_agents} -rules={n_rules} -s={is_strat}"

    start = time.time()
    os.system(full_cmd)
    end = time.time()

    return end - start


def generate_info_dump():

    info_dump = {}
    stratify = "true"
    iters = 1

    for agents in [1, 8, 16]:  # , 64]:
        for rules in [0, 10, 100]:  # , 1000]:

            total_runtime = 0
            for _ in range(iters):
                total_runtime += run(agents, rules, stratify)
            avg_runtime = total_runtime / iters

            info_dump[(agents, rules)] = avg_runtime

    return info_dump


def plotter():
    info_dump = {
        (1, 0): 0.38533611297607423,
        (1, 10): 0.3812422275543213,
        (1, 100): 0.4310877323150635,
        (8, 0): 0.5403193950653076,
        (8, 10): 0.5911492347717285,
        (8, 100): 1.1235480785369873,
        (16, 0): 1.2290499210357666,
        (16, 10): 1.511897611618042,
        (16, 100): 3.8333714485168455,
    }

    _x = np.arange(3)
    _y = np.arange(3)
    _xx, _yy = np.meshgrid(_x - 0.5, _y - 0.5)
    x, y = _xx.ravel(), _yy.ravel()

    top_ = list(info_dump.values())
    for i in range(len(top_)):
        top_[i] *= 10
    top = top_
    bottom = np.zeros_like(top)
    width = depth = 1

    print(top)
    print(bottom)
    print(x, y, x + y)

    fig = plt.figure(figsize=(8, 3))
    ax1 = fig.add_subplot(111, projection="3d")
    ax1.bar3d(x, y, bottom, width, depth, top, shade=True)
    ax1.set_title("Runtime per iteration for stratified ruleset")
    plt.xticks(_x, [0, 10, 100])
    plt.xlabel("Number of active rules")
    plt.yticks(_y, [1, 8, 16])
    plt.ylabel("Number of agents")
    ax1.set_zlabel("Runtime (ms)")
    plt.show()


plotter()
