"""
Abstract class for calculating the quantities related to the theory of
causal emergence, as described in:

Rosas FE*, Mediano PAM*, Jensen HJ, Seth AK, Barrett AB, Carhart-Harris RL, et
al. (2020) Reconciling emergences: An information-theoretic approach to
identify causal emergence in multivariate data. PLoS Comput Biol 16(12):
e1008289.
"""
from abc import ABCMeta, abstractmethod
import itertools as it
import logging
import math
import numpy as np
import pickle
from typing import Any, Callable, Dict, List, Union

import emergence.utils.log


class EmergenceCalc(metaclass = ABCMeta):
    """
    Computes quanities related to causal emergence using a given MutualInfo
    calculator function and some data describing a system of micro variables X
    (sources) and macro feature V (target)
    """
    @abstractmethod
    def __init__(self,
            X: np.ndarray, V: np.ndarray, mutualInfo: Callable,
            pointwise: bool = False, dt: int = 1, filename: str = ''
        ) -> None:
        """
        Params
        ------
        X
            system micro variables corresponding to the 'parts' or 'sources'
        V
            system macro variable of corresponding to the 'whole' or 'target'
        mutualInfo
            mutual information function
        pointwise
            whether to use pointwise (p log p) or Shannon (sum p log p) MI
        dt
            number of time steps in the future to predict
        filename
            if set, save the object to a file
        """
        logging.info(f"Initialised Emergence Calculator for system of {self.n} variables")
        logging.info(f"   with time delay {dt} between sources and target")
        logging.info(f"   using mutual information function {mutualInfo.__name__}")

        if filename:
            logging.info(f"Dumping EmergenceCalc object with all pairwise MI to {filename}_calc.pkl")
            with open(f"{filename}_calc.pkl", 'wb') as f:
                pickle.dump(self, f, pickle.HIGHEST_PROTOCOL)

        logging.info('Done.')


    @classmethod
    def correction_coef(cls, n: int, q: int, r: int) -> int:
        """
        Compute coefficient for the double-counting redundancy correction

        Params
        ------
        n
            system size i.e. number of microscopic features (sources)
        q
            order of correction, strictly smaller than n
        r
            size of the set of sources being considered in the redundancy
            calculation
        """
        if q > n:
            err = f"Order of correction q={q} must be smaller than number of sources n={n}"
            logging.error(err)
            raise ValueError(err)
        if r > n or r < n - q + 1:
            err = f"Atom set size r={r} must be strictly between {n - q + 1} and {n}"
            logging.error(err)
            raise ValueError(err)

        if r == n - q + 1:
            return n - q
        else:
            return r - 1 - sum( cls.correction_coef(n, q, s) * math.comb(r, s)
                                for s in range(n - q + 1, r) )


    def _intersection_info(self, calcs: Dict[Any, Any], ixs: List[int]) -> float:
        """
        Compute intersection information of all terms with indices in the given
        list, i.e. the minimum mutual info

        Params
        ------
        calcs
            list of MI calculators to select from
        ixs
            list of indices representing information atoms
        """
        infos = [ calcs[i] for i in ixs ]
        return min(infos)


    def _lattice_expansion(self,
            calcs: Dict[float, float], n: int, q: int = 0
        ) -> List[int]:
        """
        Expand PID latice for qth order correction.
        Given a system of size n (i.e. n sources), corrections are supported
        up to order q=n-1. The correction consists of adding (and if needed,
        subtracting) the intersection information from expanding the lattice.
        The intersection (redundant) info is defined as minimal mutual info (MMI)

        - The 1st order correction adds the MMI of all n redundant atoms
        - The 2nd order correction adds a sum of MMIs over sets of n-1 redundant
        atoms and subtracts the MMI of all n redundant atoms
        - The qth order correction adds sums over MMI over sets of n-q+1 redundant
        atoms and subtracts the MMI over sets of n-q+2 atoms, and so on

        Params
        ------
        calcs
            list of MI calculators to select from
        n
            number of sources
        q
            order of correction
        """
        corr = 0

        for r in range(n - q + 1, n + 1):
            logging.info(f"Computing correction for atom sets of size {r}")
            atom_sets = list(it.combinations(range(n), r))
            coef = self.correction_coef(n, q, r)
            logging.info(f"with {len(atom_sets)} sets with coefficient {coef}")

            corr += sum(
                coef * self._intersection_info(calcs, ixs) for ixs in atom_sets )

        return corr


    def psi(self, q: int = 0) -> Union[float, List[float]]:
        """
        Compute the Psi measure as the difference between how the sources jointly
        and individually predict the target.

            Psi = I(V(t); V(t')) - sum_i I(X_i(t); V(t')) + correction

        where the correction refers to adding back double counted redundancy in
        the I(X(t); V(t')) term, and t' - t = self.dt

        Params
        ------
        q
            compute lattice correction of order q, i.e. add back the partial
            mutual information corresponding to the q lowest levels of the PID
            lattice

        Returns
        ------
        psi
            the synergy-redundancy index with redundancy correction of order q
        """
        msg = "Computing Psi"
        if q:
            msg += f" using redundancy correction of order {q}"
        logging.info(msg)
        if q > self.n:
            err = f"Order of correction q={q} must be smaller than number of sources n={self.n}"
            logging.error(err)
            raise ValueError(err)
        syn  = self.vmiCalc
        red  = sum(xvmi for xvmi in self.xvmiCalcs.values())
        corr = self._lattice_expansion(self.xvmiCalcs, self.n, q)
        return syn - red + corr


    def gamma(self) -> Union[float, List[float]]:
        """
        Compute the causal decoupling Gamma measure as the maximum mutual info
        between the target and each individual source

            Gamma = max_j I(V(t); X_j(t'))

        where  t' - t = self.dt
        """
        gamma = max(self.vxmiCalcs.values())
        return gamma


    def delta(self, q: int = 0) -> Union[float, List[float]]:
        """
        Compute the downward causation Delta measure as the maximum difference
        between the mutual information between the target and each source and
        the sum of mutual information between that source and all other sources

            Delta = max_j (I(V(t);X_j(t')) - sum_i I(X_i(t); X_j(t') + correction)

        where the correction refers to adding back double counted redundancy in
        the I(X(t); X_j(t')) term, and t' - t = self.dt
        """
        msg = "Computing Delta"
        if q:
            msg += f" using redundancy correction of order {q}"
        logging.info(msg)
        if q > self.n:
            err = f"Order of correction q={q} must be smaller than number of sources n={self.n}"
            logging.error(err)
            raise ValueError(err)

        xmiCalcs = lambda j: { i: self.xmiCalcs[(i, j)] for i in range(self.n) }
        delta = max(
            vx - sum(xmiCalcs(j).values()) + self._lattice_expansion(xmiCalcs(j), self.n, q)
            for j, vx in enumerate(self.vxmiCalcs.values()) )
        return delta


