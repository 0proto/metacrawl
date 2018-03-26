## Metacrawl

Small service for crawling meta information from the websites.

--- 

### Usage

```
POST /tasks/
```
in `POST` body there sould be URLs separated by `\n`

Possible response codes:
```
201: Task created
400: URL list is empty
```
Response body: taskID in JSON format.

There is a strong requirement for crawling no more than 1 URL from the 1 domain per second.

HTML gets parsed via HTML Tokenized from the standard library.

To get the task by its ID:
```
GET /tasks/<taskId>/
GET parameters (optional):
delete=1 — send the CSV back and delete the task
```

Possible response codes:
```
200: task completed
204: task in progress
404: task not found
```

If task is completed, it should return the CSV with the following fields:

```
HTTP Status Code: 0, if there was a network error, -1, if page address is not a valid URL (for example, http:/ya.ru instead of http://ya.ru)
URL: page address
Page Title
Meta Description
Meta Keywords
Og:image
```

---
## Known issues

In the current implementation there is a memory leak in the domain's rate limiter. I can fix this if you want me to do so.