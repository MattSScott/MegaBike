import numpy as np
import json


def parse_full(filename):
    with open(filename) as json_file:
        parsed_json = json.load(json_file)
        game_data = parsed_json["iteration"]

        full_data_agent, full_data_bike = parse_iteration_data(game_data[0])
        for i in range(1, len(game_data)):
            new_agent, new_bike = parse_iteration_data(game_data[i])
            full_data_agent = np.append(full_data_agent, new_agent, axis=1)
            full_data_bike = np.append(full_data_bike, new_bike, axis=0)
        print(full_data_agent.shape)
        print(full_data_bike.shape)


def parse_iteration_data(iteration_data_full):
    iteration_data = iteration_data_full["round"]
    iteration_agent_data, iteration_bike_data = parse_round_data(iteration_data[0])
    for i in range(1, len(iteration_data)):
        new_agent, new_bike = parse_round_data(iteration_data[i])
        iteration_agent_data = np.append(iteration_agent_data, new_agent, axis=1)
        iteration_bike_data = np.append(iteration_bike_data, new_bike, axis=0)
    return (iteration_agent_data, iteration_bike_data)


def parse_round_data(round_data):
    bikes = round_data["bikes"]
    # given M bikes, with N agents per bike...
    # bike = 1xM, agents = N*M
    round_bike_dirs = np.empty((len(bikes), 2))
    round_agent_dirs = np.empty((8, len(bikes), 2))
    round_bike_dirs.fill(np.nan)
    round_agent_dirs.fill(np.nan)
    for idx, bike in enumerate(bikes.values()):
        dir = bike["bikeDirection"]
        bike_ag_dir = parse_bike_data(bike)
        round_agent_dirs[:, idx, :] = bike_ag_dir
        round_bike_dirs[idx, 0] = dir["x"]
        round_bike_dirs[idx, 1] = dir["y"]
    return (round_agent_dirs, round_bike_dirs)


cnt = 0


def parse_bike_data(bike_data):
    agents = bike_data["agents"]
    # print(len(agents))
    global cnt
    if len(agents) != 0:
        cnt += 1
    agent_dirs = np.empty((8, 2))
    agent_dirs.fill(np.nan)
    for idx, agent in enumerate(agents.values()):
        dir = agent["agentDirection"]
        agent_dirs[idx, 0] = dir["x"]
        agent_dirs[idx, 1] = dir["y"]

    return agent_dirs


parse_full("gameDumps/debug/6142fc5c-a09c-4a9a-9d10-4ff496a8f96e.json")
# parse_full("gameDumps/debug/76e34653-faaf-4805-b16c-6c8b8eb362ad.json")
print(cnt)
