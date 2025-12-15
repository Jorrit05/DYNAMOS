# import pandas as pd
#
# # Example DataFrame (can be empty or have other columns)
# data_df = pd.DataFrame({})
#
# # New trace value to add
# new_trace_value = "ml-merge"
#
# # Check if "trace" is a column
# if "trace" in data_df.columns:
#     # Append a new row with only "trace" value
#     new_row = pd.DataFrame([{"trace": new_trace_value}])
#     data_df = pd.concat([data_df, new_row], ignore_index=True)
# else:
#     # Create a new DataFrame with only the "trace" column
#     data_df = pd.DataFrame([{"trace": new_trace_value}])
#
# print(data_df)
#

import pandas as pd
import numpy as np
from diffprivlib.mechanisms import Laplace

# diffprivlib==0.6.6

def add_laplace_noise_dp(df: pd.DataFrame, epsilon: float = 1.0, sensitivity: float = 1.0) -> pd.DataFrame:
    """
    Adds Laplace noise using diffprivlib for differential privacy.
    :param df: pandas DataFrame
    :param epsilon: privacy parameter (smaller = more noise)
    :param sensitivity: sensitivity of the query (max value change by changing a single row)
    :return: new DataFrame with noisy columns
    """
    try:
        numerical_cols = df.select_dtypes(include='number').columns.tolist()
        if "building_id" in numerical_cols:
            numerical_cols.remove("building_id")

        noisy_df = df.copy()
        scale = sensitivity / epsilon

        for col in numerical_cols:
            mechanism = Laplace(epsilon=epsilon, delta=0.0, sensitivity=sensitivity)
            noisy_values = []
            for v in noisy_df[col]:
                noisy_values.append(mechanism.randomise(v))
            noisy_df[col] = noisy_values

        return noisy_df
    except Exception as e:
        print(f"Error adding Laplace noise with diffprivlib: {e}")
        return pd.DataFrame()


df = pd.DataFrame({
    "building_id": [101, 102, 103, 104],
    "temperature": [20.5, 21.0, 19.8, 22.1],
    "energy_usage": [150.0, 160.5, 155.2, 162.3],
    "occupancy": [35, 40, 38, 37]
})

df_noisy = add_laplace_noise_dp(df, epsilon=1.0, sensitivity=1.0)


print("Original DataFrame:")
print(df)
print("\nNoisy DataFrame:")
print(df_noisy)

# --------------

import pandas as pd

weather_df = pd.read_csv(r"./python/ml-merge-datasets/data/yearly_weather.csv")
building_df = pd.read_csv(r"./python/ml-merge-datasets/data/yearly_building.csv")




weather_df.info()
print(*list(weather_df.columns), sep="\n")
weather_df.describe()
weather_df.head()

col = "air_temperature"
for col in weather_df.columns:
    print(col)
    print("\t".join([ str(round(x,2)) for x in  list(weather_df[col].describe().loc[['min', 'mean', 'max', 'std']].values)]))


print(*list(building_df.columns), sep="\n")
building_df.info()
building_df.describe()
building_df.head()


for col in building_df.columns:
    print(col)
    try:
        print("\t".join([ str(round(x,2)) for x in  list(building_df[col].describe().loc[['min', 'mean', 'max', 'std']].values)]))
    except Exception as e:
        print(e)

building_df['primary_use'].value_counts()




# values = [1.75,	1.254	,0.534]
# values = [1.728	,2.066,	0.998]
# values = [1.616,	2.534,	1.242]
# values = [1.656,	1.1614,	1.446]
# values = [1.8378,	1.7314,	2.666]
values = [1.6132,	2.3152,	3.92]
total = sum(values)

percentages = [(v / total) * 100 for v in values]
print(percentages)  # Output: [20.0, 20.0, 60.0]
print(*percentages, sep="\t")



s = """
12
7
9
8
15
7
15
4
60
3
2
1
6
2
4
19
76
6
2
2
35
8
30
1
40
4
8
15
4
11
14
2
15
70
80
"""

sum([float(i) for i in s.split()])
