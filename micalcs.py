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

def handle_nans(X, V, drop = True):
    x_nan_ind = np.any(~np.any(np.isnan(X), axis=1), axis = 1)
    v_nan_ind = ~np.any(np.isnan(V), axis = 1)
    nan_ind = np.all(np.array([x_nan_ind, v_nan_ind]).T, axis = 1)
    if drop:
        return X[nan_ind, :], V[nan_ind, :]
    else:
        for i,n in enumerate(nan_ind):
            X[i] = np.nan_to_num(X[i], nan=np.nanmean(X[i]))
            V[i] = np.nan_to_num(V[i], nan=np.nanmean(V[i]))
        return X, V

def load_run(path, d, f, calctype = 'Gaussian'):
    print(f"Loading run {f}")
    with open(f"{path}{d}/{f}/bike_data.pickle", 'rb') as dat:
        bike = pickle.load(dat)
        print(f"Loaded bike with shape {bike.shape}")
    with open(f"{path}{d}/{f}/agent_data.pickle", 'rb') as dat:
        agents = pickle.load(dat)
        print(f"Loaded agents with shape {agents.shape}")

    N, T, D = agents.shape
    #X = np.nan_to_num(agents.transpose(1, 0, 2), -100)
    #V = np.nan_to_num(bike, -100)
    X = agents.transpose(1, 0, 2)
    V = bike
    X, V = handle_nans(X, V)
    print(f"Kept agent/bike data with shape {X.shape}")
    calc = JidtCalc(X, V, MutualInfo.get(calctype), pointwise = False, dt = 1,
                    filename = f"../MegaBike/MICalcs/{d}/{calctype}/{f}")


def run_calcs(d, calctype):
    JVM.start()
    runs = lambda d: os.listdir(f"{path}/{d}")
    for f in runs(d):
        load_run(path, d, f, calctype)
    #JVM.stop()


def plot_agent_nets(calc, figpath, ax = plt.gca(), viz_scale = 10):
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
    #TODO: nice colourmap
    #cmaps = plt.cm.ScalarMappable(norm = plt.Normalize(
    #    vmin = min(widths.values()), vmax = max(widths.values())))
    #plt.colorbar(cmaps, cax = ax)

    plt.savefig(figpath, dpi = 300)


def agg_calcs(d, calctype):
    JVM.start()

    calc = None
    emergence = []

    cpath = f"../MegaBike/MICalcs/{d}/"
    cfiles = [ f for f in os.listdir(f"{cpath}{calctype}/") if 'pkl' in f ]
    for cfilename in cfiles:
        with open(f"{cpath}{calctype}/{cfilename}", 'rb') as f:
            calc = pickle.load(f)
        #plot_agent_nets(calc, f"{cpath}{calctype}/{cfilename.split('.')[0]}.png")
        emergence.append([ cfilename.split('_')[0]
                         , calc.psi(q = 7)
                         , calc.gamma()
                         , calc.delta(q = 7) ])

    df = pd.DataFrame(index = range(1, 31), columns = ['Hash', 'Psi', 'Gamma', 'Delta' ], data = emergence)
    df.to_csv(f"{cpath}/{d}_{calctype}_emergence_criteria.csv")
    JVM.stop()


if __name__ == "__main__":
    for estimator in [ 'Kraskov1' ]: #, 'Kernel', 'Gaussian' ]:
        run_calcs('mutable', estimator)
        agg_calcs('mutable', estimator)

        run_calcs('immutable', estimator)
        agg_calcs('immutable', estimator)
