from itertools import count
from typing import Literal

from fastapi import FastAPI, HTTPException
from fastapi.responses import HTMLResponse
from pydantic import BaseModel


app = FastAPI(title="Todo App")
todo_id = count(1)
todos = [
    {"id": next(todo_id), "title": "Buy milk", "completed": False},
    {"id": next(todo_id), "title": "Write some code", "completed": True},
]


class TodoCreate(BaseModel):
    title: str


class TodoUpdate(BaseModel):
    completed: bool


class TodoItem(BaseModel):
    id: int
    title: str
    completed: bool


@app.get("/", response_class=HTMLResponse)
def index() -> str:
    return """
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>FastAPI Todo</title>
  <style>
    :root {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: #1f2937;
      background: #f3f4f6;
    }
    body {
      margin: 0;
      padding: 32px 16px;
    }
    main {
      max-width: 640px;
      margin: 0 auto;
      background: white;
      border-radius: 16px;
      padding: 24px;
      box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
    }
    form, li {
      display: flex;
      gap: 12px;
      align-items: center;
    }
    ul {
      list-style: none;
      padding: 0;
    }
    li {
      justify-content: space-between;
      padding: 12px 0;
      border-bottom: 1px solid #e5e7eb;
    }
    button {
      border: 0;
      border-radius: 10px;
      padding: 10px 14px;
      cursor: pointer;
      background: #111827;
      color: white;
    }
    input[type="text"] {
      flex: 1;
      padding: 10px 12px;
      border-radius: 10px;
      border: 1px solid #d1d5db;
    }
    .done {
      text-decoration: line-through;
      color: #6b7280;
    }
  </style>
</head>
<body>
  <main>
    <h1>FastAPI Todo</h1>
    <form id="todo-form">
      <input id="title" type="text" placeholder="Add a task" required>
      <button type="submit">Add</button>
    </form>
    <ul id="list"></ul>
  </main>
  <script>
    async function loadTodos() {
      const response = await fetch('/api/todos');
      const todos = await response.json();
      const list = document.getElementById('list');
      list.innerHTML = '';
      for (const todo of todos) {
        const item = document.createElement('li');
        const left = document.createElement('label');
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.checked = todo.completed;
        checkbox.onchange = async () => {
          await fetch(`/api/todos/${todo.id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed: checkbox.checked })
          });
          loadTodos();
        };
        const text = document.createElement('span');
        text.textContent = todo.title;
        if (todo.completed) text.className = 'done';
        left.append(checkbox, text);

        const remove = document.createElement('button');
        remove.textContent = 'Delete';
        remove.onclick = async () => {
          await fetch(`/api/todos/${todo.id}`, { method: 'DELETE' });
          loadTodos();
        };
        item.append(left, remove);
        list.append(item);
      }
    }

    document.getElementById('todo-form').onsubmit = async (event) => {
      event.preventDefault();
      const input = document.getElementById('title');
      await fetch('/api/todos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: input.value })
      });
      input.value = '';
      loadTodos();
    };

    loadTodos();
  </script>
</body>
</html>
"""


@app.get("/api/todos", response_model=list[TodoItem])
def list_todos() -> list[dict]:
    return todos


@app.post("/api/todos", response_model=TodoItem)
def create_todo(payload: TodoCreate) -> dict:
    item = {"id": next(todo_id), "title": payload.title.strip(), "completed": False}
    todos.append(item)
    return item


@app.patch("/api/todos/{item_id}", response_model=TodoItem)
def update_todo(item_id: int, payload: TodoUpdate) -> dict:
    for todo in todos:
        if todo["id"] == item_id:
            todo["completed"] = payload.completed
            return todo
    raise HTTPException(status_code=404, detail="Todo not found")


@app.delete("/api/todos/{item_id}", response_model=dict[str, Literal[True]])
def delete_todo(item_id: int) -> dict[str, Literal[True]]:
    for index, todo in enumerate(todos):
        if todo["id"] == item_id:
            todos.pop(index)
            return {"ok": True}
    raise HTTPException(status_code=404, detail="Todo not found")
