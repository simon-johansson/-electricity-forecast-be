CREATE TABLE country_data
(
    countryName TEXT NOT NULL PRIMARY KEY,
    json        TEXT NOT NULL
);

CREATE TABLE available_countries
(
    id   TEXT NOT NULL PRIMARY KEY,
    json TEXT NOT NULL
);
