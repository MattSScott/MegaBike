from dit.distribution import BaseDistribution
import logging
import numpy as np
from typing import Callable, Dict, List, Tuple, Union, Set

import emergence.utils.log
from emergence.calc.base import EmergenceCalc


class DitDistCalc(EmergenceCalc):
    """
    The class below uses the Shannon mutual information calculator applied to
    distributions of discrete variables describing a system of source variables
    X and a target macroscopic feature V.

    Uses DIT distributions and Shannon mutual information.
    """
    def __init__(self,
            dist: BaseDistribution, mutualInfo: Callable,
            filename: str = ''
        ) -> None:
        """
        Initialise class with the joint distribution corresponding to the system
        variables. The last random variable in the distribution, by convention,
        is the target (macro) variable. Then compute all the pairwise mutual
        informations.

        Params
        ------
        dist
            joint probability distribution of the system source and target vars
        mutualInfo
            mutual information function, Shannon
        filename
            if set, save the object to a file
        """
        self.X = dist.rvs[:-1]
        self.V = dist.rvs[-1]
        self.n = len(self.X)

        logging.info(f"Computing mutual informations: MI(V[t], V[t])")
        self.vmiCalc = mutualInfo(dist, [ x[0] for x in self.X ], self.V)

        logging.info(f"Computing mutual informations: MI(Xi[t], V[t])")
        self.xvmiCalcs = { i: mutualInfo(dist, x, self.V) for i, x in enumerate(self.X) }

        logging.info(f"Computing mutual informations: MI(V[t], Xi[t])")
        self.vxmiCalcs = { i: mutualInfo(dist, self.V, x) for i, x in enumerate(self.X) }

        logging.info(f"Computing mutual informations: MI(Xi[t], Xj[t])")
        self.xmiCalcs = dict()
        for i in range(self.n):
            for j in range(self.n):
                self.xmiCalcs[(i, j)] = mutualInfo(dist, self.X[i], self.X[j])

        super().__init__(self.X, self.V, mutualInfo, False, 0, filename)

