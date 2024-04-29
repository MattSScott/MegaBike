import matplotlib.pyplot as plt
import networkx as nx
import numpy as np
import os
import pandas as pd
import pickle

from emergence.calc.jidt import JidtCalc
from emergence.micalc import MutualInfo
from emergence.utils.jvm import JVM

path = "../MegaBike/vectorisedDumps/"

def load_run(path, d, f):
    print(f"Loading run {f}")
    with open(f"{path}{d}/{f}/bike_data.pickle", 'rb') as dat:
        bike = pickle.load(dat)
        print(f"Loaded bike with shape {bike.shape}")
    with open(f"{path}{d}/{f}/agent_data.pickle", 'rb') as dat:
        agents = pickle.load(dat)
        print(f"Loaded agents with shape {agents.shape}")

    N, T, D = agents.shape
    X = np.nan_to_num(agents.transpose(1, 0, 2), -100)
    V = np.nan_to_num(bike, -100)
    calc = JidtCalc(X, V, MutualInfo.get('Gaussian'), pointwise = False, dt = 1,
                    filename = f"../MegaBike/MICalcs/{d}/{f}")


def run_calcs(d):
    JVM.start()
    runs = lambda d: os.listdir(f"{path}/{d}")
    for f in runs(d):
        load_run(path, d, f)
    JVM.stop()


def plot_agent_nets(calc, figpath, ax = plt.gca(), viz_scale = 100):
    net = nx.DiGraph()
    edges = [ (u+1, v+1, w) for (u,v),w in calc.xmiCalcs.items() ]
    net.add_weighted_edges_from(edges)
    nodelist = net.nodes()
    widths = nx.get_edge_attributes(net, 'weight')
    ax.axis('square')

    pos = nx.get_node_attributes(net, 'pos')
    if not pos:
        pos = nx.shell_layout(net)
    nx.draw_networkx_nodes(net, pos, ax = ax,
                           nodelist = nodelist,
                           node_size = 500,
                           node_color = 'black',
                           alpha = 0.7)
    collection = nx.draw_networkx_edges(net, pos, ax = ax,
                           edgelist = widths.keys(),
                           width = np.array(list(widths.values())) * viz_scale,
                           edge_color = list(widths.values()),
                           alpha = 1)
    nx.draw_networkx_labels(net, pos = pos, ax = ax,
                            labels = dict(zip(nodelist,nodelist)),
                            font_color = 'white')
    #cmaps = plt.cm.ScalarMappable(norm = plt.Normalize(
    #    vmin = min(widths.values()), vmax = max(widths.values())))
    #plt.colorbar(cmaps, cax = ax)

    plt.savefig(figpath, dpi = 300)


def agg_calcs(d):
    cpath = f"../MegaBike/MICalcs/{d}"
    JVM.start()

    calc = None
    emergence = []

    cfiles = [ f for f in os.listdir(f"{cpath}") if 'pkl' in f ]
    for cfilename in cfiles:
        with open(f"{cpath}/{cfilename}", 'rb') as f:
            calc = pickle.load(f)
        plot_agent_nets(calc, f"{cpath}/{cfilename.split('.')[0]}.png")
        emergence.append([ calc.psi(q = 7)
                         , calc.gamma()
                         , calc.delta(q = 7) ])

    emergence = np.array(emergence)
    df = pd.DataFrame(index = range(1, 31), columns = ['Psi', 'Gamma', 'Delta' ], data = emergence)
    df.to_csv(f"{cpath}/{d}_emergence_criteria.csv")

    JVM.stop()


run_calcs('immutable')
agg_calcs('immutable')
