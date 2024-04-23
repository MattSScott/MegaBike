import numpy as np
import json


def parse_full(filename):
    with open(filename) as json_file:
        parsed_json = json.load(json_file)
        iters = parsed_json["iteration"]
        # print(iters[0].keys())
        parse_iteration_data(iters[0]["round"])


def parse_iteration_data(iteration_data):
    parse_round_data(iteration_data[0])


def parse_round_data(round_data):
    bikes = round_data["bikes"]
    # given M bikes, with N agents per bike...
    # bike = 1xM, agents = N*M
    bike_dirs = []
    agent_dirs = []
    for bike in bikes.values():
        # print("BD:", bike["bikeDirection"])
        dir = bike["bikeDirection"]
        bike_ag_dir = parse_bike_data(bike)
        agent_dirs.append(bike_ag_dir)
        bike_dirs.append([dir["x"], dir["y"]])
    # print(bike_dirs)
    # print()
    # print(len(agent_dirs))
    # for d in agent_dirs:
    #     print(len(d))
    #     for y in d:
    #         print(len(y))

    # a_arr = np.array(agent_dirs)


def parse_bike_data(bike_data):
    agents = bike_data["agents"]
    print(len(agents))
    agent_dirs = []
    for agent in agents.values():
        dir = agent["agentDirection"]
        agent_dirs.append([dir["x"], dir["y"]])
    return agent_dirs


# parse_full("gameDumps/mutable/00cd13d1-4d35-4667-aa9d-87c32f2d3eaf.json")
parse_full("gameDumps/debug/76e34653-faaf-4805-b16c-6c8b8eb362ad.json")
