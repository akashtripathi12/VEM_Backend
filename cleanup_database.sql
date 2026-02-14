-- SQL Queries to Filter Database to 10 Countries

-- Step 1: Delete cities that are NOT from the 10 target countries
DELETE FROM cities 
WHERE country_code NOT IN ('US', 'DE', 'SG', 'AE', 'GB', 'ES', 'TH', 'CN', 'AU', 'IN');

-- Step 2: Delete countries that are NOT in the 10 target countries
DELETE FROM countries 
WHERE code NOT IN ('US', 'DE', 'SG', 'AE', 'GB', 'ES', 'TH', 'IN');

-- Verification Queries:

-- Check remaining countries (should be 10)
SELECT COUNT(*) as total_countries FROM countries;
SELECT code, name FROM countries ORDER BY code;

-- Check remaining cities
SELECT COUNT(*) as total_cities FROM cities;

-- Check cities per country
SELECT country_code, COUNT(*) as city_count 
FROM cities 
GROUP BY country_code 
ORDER BY country_code;
