package main

const (
	PlanningResponseFormat = `<response_format>
{
  "_thinking": "explanation of your interpretation and decision process",
  "task": "task_name",
  "content" : "what should next step focus on"
}
</response_format>`
	PlanningSystemPrompt = `<objective>
You are a task dispatcher service. Based on inputted message and context
</objective>
<rules>
- ALWAYS output a valid JSON object with "_thinking", "task" and "content" properties
- ALWAYS output in english
- Available rest APIs: places, people and gps
- GPS api only works with userId field, if there is no userIds in either context or message DO NOT use this api. It only accept 1 userId.
- RELATION api is only for relation between people. Not between places and people
- PEOPLE api - only know all the places queried person was in. It only accepts 1 name.
- PLACES api - only knows which people were at the specific place. It only accepts 1 place name.
- DB - knows: connections between people, datacenters and about users (like userIds). It's a SQL database
- if you are querying specific API only respond with the relevant information
- If you query about RAFAŁ as input, convert the name to RAFAL before querying. Except for the SQL Database
- userIds column in the user table is named id
- Only query the database for one user at a time
- for the GPS API, only accepts userID as input, nothing else
- the "content" field is always a string. Do not put json object there
- If you have final answer put the result in "content" field and set the task to "COMPLETE"
- in the final answer, include names in the same format as in the input
</rules>
<task_list>
- QUERY_DB - send query to sql db api
- QUERY_API API_NAME- send query to REST API. API_NAMEs: places, people or gps 
- RELATION - send query to neo4j service agent
- COMPLETE - task is completed
</task_list>
<examples>
USER: I want to know who knows Rafał.
AI:
{
  "_thinking": "The user asked for people who know Rafał. This implies a relationship network query, which is suitable for a neo4j database that handles connections between entities such as people.",
  "task": "RELATION",
  "content": "Identifying all individuals who have a connection with Rafał."
}
USER: I want to know who was in Grudziądz.
AI:
{
  "_thinking": "The user is asking for information about people who were in the location called Grudziądz. To find this information, I need to use the PLACES API because it provides details about which people were at specific places.",
  "task": "QUERY_API PLACES",
  "content": "GRUDZIADZ"
}
USER: I want to know every place Azazel was in.
AI:
{
  "_thinking": "The user is asking for information about all the places a person named Azazel has been. According to the rules, the people API can provide information on all the places a specific person was in.",
  "task": "QUERY_API PEOPLE",
  "content": "AZAZEL"
}
User: I want to know userIds for Rafał
AI:
{
  "_thinking": "The user is asking for the user IDs of Rafał. This information about users should be extracted from a database that maintains records about people and their associated user IDs.",
  "task": "QUERY_DB",
  "content": "SELECT id FROM users WHERE username = 'Rafał';"
}
User: what are the gps coordinates of user with id 28?
AI:
{
  "_thinking": "The user is asking for the coordinates of a user with a specific ID. To retrieve this, the GPS API should be utilized since it works with userId to get the location.",
  "task": "QUERY_API GPS",
  "content": "28"
}
</examples>`
	CleanUpSystemPrompt = `<objective>
Your task is to clean up the given message
</objective>
<rules>
- only reformat the message, do not change the meaning
- do not add or remove any information
- do not change the order of the information
- do not add any additional information
</rules>
<examples>
<response_format>
{
  "_thinking": "explanation of your interpretation and decision process",
  "answer" : [
{
    "imie": {
        "lat": 12.345,
        "lon": 65.431
    },
    "kolejne-imie": {
        "lat": 19.433,
        "lon": 12.123
    }
}
]
}
</response_format>`
)
