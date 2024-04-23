import os
import time
import matplotlib.pyplot as plt
import numpy as np


AGENT_ARR = [1, 8, 16, 32]
RULE_ARR = [0, 10, 100, 1000]


def run(n_agents=50, n_rules=10, is_strat="true"):

    full_cmd = f"go run . -agents={n_agents} -rules={n_rules} -s={is_strat}"

    start = time.time()
    os.system(full_cmd)
    end = time.time()

    return end - start


def generate_info_dump():

    info_dump = {}
    stratify = "false"
    iters = 30

    for agents in AGENT_ARR:
        for rules in RULE_ARR:

            total_runtime = 0
            for _ in range(iters):
                total_runtime += run(agents, rules, stratify)
            avg_runtime = total_runtime / iters

            info_dump[(agents, rules)] = avg_runtime

    return info_dump


def plotter():

    info_dump = generate_info_dump()
    print(info_dump)  # output in case of graphical error

    _x = np.arange(len(RULE_ARR))
    _y = np.arange(len(AGENT_ARR))
    _xx, _yy = np.meshgrid(_x - 0.5, _y - 0.5)
    x, y = _xx.ravel(), _yy.ravel()

    top_ = list(info_dump.values())
    for i in range(len(top_)):
        top_[i] *= 10
    top = top_
    bottom = np.zeros_like(top)
    width = depth = 1

    fig = plt.figure(figsize=(8, 3))
    ax1 = fig.add_subplot(111, projection="3d")
    ax1.bar3d(x, y, bottom, width, depth, top, shade=True)
    ax1.set_title("Runtime per iteration for stratified ruleset")
    plt.xticks(_x, RULE_ARR)
    plt.xlabel("Number of active rules")
    plt.yticks(_y, AGENT_ARR)
    plt.ylabel("Number of agents")
    ax1.set_zlabel("Runtime (ms)")
    plt.show()


def dual_plotter():
    active = {
        (1, 0): 0.5437193632125854,
        (1, 10): 0.36507601737976075,
        (1, 100): 0.40704824924468996,
        (1, 1000): 0.74881010055542,
        (8, 0): 0.5119576295216878,
        (8, 10): 0.568668254216512,
        (8, 100): 1.0277202685674032,
        (8, 1000): 5.597009364763895,
        (16, 0): 1.2146671215693157,
        (16, 10): 1.5255611737569172,
        (16, 100): 4.166511090596517,
        (16, 1000): 29.562002436319986,
        (32, 0): 5.177097964286804,
        (32, 10): 6.4118003686269125,
        (32, 100): 17.246908402442934,
        (32, 1000): 125.5293609380722,
    }

    inactive = {
        (1, 0): 0.4477794329325358,
        (1, 10): 0.3971668243408203,
        (1, 100): 0.5514362970987956,
        (1, 1000): 1.9774396101633707,
        (8, 0): 0.526043192545573,
        (8, 10): 0.7808104674021403,
        (8, 100): 3.015770610173543,
        (8, 1000): 26.68324653307597,
        (16, 0): 1.2425683736801147,
        (16, 10): 2.622816562652588,
        (16, 100): 15.013193734486897,
        (16, 1000): 136.38989001115164,
        (32, 0): 5.284009154637655,
        (32, 10): 11.451239760716756,
        (32, 100): 65.96472969055176,
        (32, 1000): 610.9927232027054,
    }

    _x = np.arange(len(RULE_ARR))
    _y = np.arange(len(AGENT_ARR))
    _xx, _yy = np.meshgrid(_x - 0.5, _y - 0.5)
    x, y = _xx.ravel(), _yy.ravel()

    top_active = list(active.values())
    for i in range(len(top_active)):
        top_active[i] *= 10
    top_inactive = list(inactive.values())
    for i in range(len(top_inactive)):
        top_inactive[i] *= 10

    bottom_active = np.zeros_like(top_active)
    bottom_inactive = np.zeros_like(top_inactive)
    width = depth = 1

    fig = plt.figure(figsize=(10, 5))
    ax1 = fig.add_subplot(121, projection="3d")
    ax2 = fig.add_subplot(122, projection="3d")
    ax1.bar3d(x, y, bottom_active, width, depth, top_active, shade=True)
    ax1.set_title("Runtime per iteration for stratified ruleset")
    ax2.bar3d(x, y, bottom_inactive, width, depth, top_inactive, shade=True)
    ax2.set_title("Runtime per iteration for non-stratified ruleset")
    for ax in [ax1, ax2]:
        ax.set_xticks(_x, RULE_ARR)
        ax.set_xlabel("Number of active rules")
        ax.set_yticks(_y, AGENT_ARR)
        ax.set_ylabel("Number of agents")
        ax.set_zlabel("Runtime (ms)")
    plt.show()


dual_plotter()
