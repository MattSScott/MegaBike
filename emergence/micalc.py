import jpype as jp
import numpy as np
from typing import Callable, Dict, Iterable, List, Tuple, Union

from emergence.utils.jvm import JVM


def _MICalcAvg(calcName: Callable[None, str]) -> Union[np.ndarray, float]:
    """
    Store the time series as observations in the mutual information calculators,
    used to estimate joint distributions between X and Y by using the pairs
    X[t], Y[t+dt], then compute the mutual information.

    Params
    ------
    X
        1st time series of shape (T, Dx) i.e. source.
        T is time or observation index, Dx is variable number.
    Y
        2nd time series of shape (T, Dy) i.e. target
        T is time or observation index, Dy is variable number.
    pointwise
        if set use pointwise MI rather than Shannon MI, i.e. applied on
        specific states rather than integrateing whole distributions.
        this returns a PMI value for each pair X[t], Y[t+dt] rather than
        a value for the whole time series
    dt
        source-destination lag

    Returns
    ------
    I(X[t], Y[t+dt]) for t in range(0, T - dt)
        if pointiwse = False, the MI is a float in nats not bits!
        if pointwise = True,  the MI is a list of floats, in nats, for every
                                pair of values in the two time series
    """
    def __compute(
            X: np.ndarray, Y: np.ndarray,
            pointwise: bool = False, dt: int = 0,
        ) -> np.ndarray:
        """
        Whenever Python decorator @_MICalcAvg is used with a function, the func
        returns the mutual info as computed with the estimator specified by the
        function.
        """
        if len(X) != len(Y):
            raise ValueError('Cannot compute MI for time series of different lengths')

        calc = jp.JClass(calcName())()

        # we displace the second array to compute time delays
        if 'Discrete' in calcName():
            jX, jY = (JVM.javify(X[:-dt], jp.JInt), JVM.javify(Y[dt:], jp.JInt))
        else:
            jX, jY = (JVM.javify(X[:-dt], jp.JDouble), JVM.javify(Y[dt:], jp.JDouble))

        # the discrete MI calc in JIDT is initialised with alphabet sizes and
        # data must be added incrementally by .addObservations(int[], int[])
        if 'Discrete' in calcName():
            params = (len(np.unique(X)), len(np.unique(Y)), 0)
            calc.initialise(*params)
            if len(X.shape) > 1 and X.shape[1] != 1 or len(Y.shape) > 1 and Y.shape[1] != 1:
                raise ValueError('Discrete calculator not supported for multivariate systems')
            calc.addObservations(jX, jY)
        # all the continuous MI calcs in JIDT are initialised with the number
        # of variables and data can be added also all at once
        else:
            calc.initialise(X.shape[1], Y.shape[1])
            calc.setObservations(jX, jY)
            calc.finaliseAddObservations()

        #TODO: doesnt work
        #calc.setProperty('PROP_TIME_DIFF', str(dt))

        if pointwise:
            # type JArray, e.g. <class 'jpype._jarray.double[]'>, can be indexed with arr[i]
            return calc.computeLocalUsingPreviousObservations(jX, jY)
        else:
            # float
            return calc.computeAverageLocalOfObservations()

    # Set name and docstrings of decorated function
    __compute.__name__ = calcName.__name__
    __compute.__doc__  = calcName.__doc__ + _MICalcAvg.__doc__
    return __compute


class MutualInfo:
    """
    Class for calling various mutual information calculators for. Returns class
    functions that can be passed to EmergenceCalc to compute mutual information.
    """
    @classmethod
    def get(self, name: str) -> Callable[None, str]:
        if name.lower() == 'discrete':
            return self.Discrete
        elif name.lower() == 'gaussian':
            return self.ContinuousGaussian
        elif name.lower() == 'kraskov1':
            return self.ContinuousKraskov1
        elif name.lower() == 'kraskov2':
            return self.ContinuousKraskov2
        elif name.lower() == 'kernel':
            return self.ContinuousKernel
        else:
            raise ValueError(f"Estimator {name} not supported")


    @_MICalcAvg
    def Discrete() -> str:
        """
        Compute discrete mutual information using Shannon entropy between time
        series X and Y.
        """
        return 'infodynamics.measures.discrete.MutualInformationCalculatorDiscrete'

    @_MICalcAvg
    def ContinuousGaussian() -> str:
        """
        Compute continuous mutual information (using differential entropy instead
        of Shannon entropy) between time series X and Y.
        The estimator assumes that the underlying distributions for the time
        series are Gaussian and that the time series are stationary.
        """
        return 'infodynamics.measures.continuous.gaussian.MutualInfoCalculatorMultiVariateGaussian'

    @_MICalcAvg
    def ContinuousKraskov1() -> str:
        """
        Compute continuous mutual information using the Kraskov estimator between
        time series X and Y, specifically the 1nd algorithm in:

        Kraskov, A., Stoegbauer, H., Grassberger, P., "Estimating mutual information",
        Physical Review E 69, (2004) 066138.

        The mutual info calculator makes no Gaussian assumptions, but the time series
        must be stationary.
        """
        return 'infodynamics.measures.continuous.kraskov.MutualInfoCalculatorMultiVariateKraskov1'

    @_MICalcAvg
    def ContinuousKraskov2() -> str:
        """
        Compute continuous mutual information using the Kraskov estimator between
        time series X and Y, specifically the 2nd algorithm in:

        Kraskov, A., Stoegbauer, H., Grassberger, P., "Estimating mutual information",
        Physical Review E 69, (2004) 066138.

        The mutual info calculator makes no Gaussian assumptions, but the time series
        must be stationary.
        """
        return 'infodynamics.measures.continuous.kraskov.MutualInfoCalculatorMultiVariateKraskov2'

    @_MICalcAvg
    def ContinuousKernel() -> str:
        """
        Compute continuous mutual information using box-kernel estimation between
        time series X and Y.
        The mutual info calculator makes no Gaussian assumptions, but the time series
        must be stationary.
        """
        return 'infodynamics.measures.continuous.kernel.MutualInfoCalculatorMultiVariateKernel'

