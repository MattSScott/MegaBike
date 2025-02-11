import logging
import numpy as np
from typing import Callable

from emergence.calc.base import EmergenceCalc


class JidtCalc(EmergenceCalc):
    """
    The class below supports a number of mutual information estimators which
    are defined in `micalc.py`, which are applied to time series describing a
    system of source variabes X and a target macroscopic feature V.

    Uses JIDT mutual information estimators Discrete, Gaussian, Kernel, Kraskov.
    """
    def __init__(self,
            X: np.ndarray, V: np.ndarray,
            mutualInfo: Callable[[np.ndarray, np.ndarray, bool, int], float],
            pointwise: bool = False,
            dt: int =  1, filename: str = ''
        ) -> None:
        """
        Initialise class with the time series corresponding to the micro Xs and
        the macro V, set any properties for the computation, then proceed to
        compute all pairwise mutual info where data is available. By default, works
        with multivariate systems, so if any txn 2D array of t states and n scalar
        variable is passed, reshape the array to (t, n, 1) first.

        Params
        ------
        X
            system micro variables of shape (t, n, d) for n d-dimensional time
            series corresponding to the 'parts' in the system
        V
            candidate emergence feature: d-dimensional system macro variable of
            shape (t, d)
        mutualInfo
            mutual information function to use from MutualInfo class
        pointwise
            whether to use pointwise (p log p) or Shannon (sum p log p) MI
        dt
            number of time steps in the future to predict
        filename
            if set, save the object to a file
        """
        if len(X.shape) < 3:
            X = np.atleast_3d(X)
        if len(V.shape) < 2:
            V = V[:, np.newaxis]

        (t, n, d) = X.shape
        self.n = n
        self.V = V
        self.X = [ X[:, i] for i in range(n) ]

        logging.info(f"Computing mutual informations: MI(V[t], V[t+{dt}])")
        self.vmiCalc = mutualInfo(V,  V,  pointwise, dt)

        logging.info(f"Computing mutual informations: MI(Xi[t], V[t+{dt}])")
        self.xvmiCalcs = dict()
        for i in range(n):
            self.xvmiCalcs[i] = mutualInfo(self.X[i], V, pointwise, dt)

        logging.info(f"Computing mutual informations: MI(V[t], Xi[t+{dt}])")
        self.vxmiCalcs = dict()
        for i in range(n):
            self.vxmiCalcs[i] = mutualInfo(V, self.X[i], pointwise, dt)

        logging.info(f"Computing mutual informations: MI(Xi[t], Xj[t+{dt}])")
        self.xmiCalcs = dict()
        for i in range(n):
            for j in range(n):
                self.xmiCalcs[(i, j)] = mutualInfo(self.X[i], self.X[j], pointwise, dt)

        super().__init__(X, V, mutualInfo, pointwise, dt, filename)

