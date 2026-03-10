const express = require("express");

const app = express();
const port = 3000;

let nextId = 3;
const todos = [
  { id: 1, title: "Buy milk", completed: false },
  { id: 2, title: "Write some code", completed: true },
];

app.use(express.json());

app.get("/", (_req, res) => {
  res.type("html").send(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Express Todo</title>
  <style>
    :root { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; color: #1f2937; background: #f3f4f6; }
    body { margin: 0; padding: 32px 16px; }
    main { max-width: 640px; margin: 0 auto; background: white; border-radius: 16px; padding: 24px; box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08); }
    form, li { display: flex; gap: 12px; align-items: center; }
    ul { list-style: none; padding: 0; }
    li { justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #e5e7eb; }
    button { border: 0; border-radius: 10px; padding: 10px 14px; cursor: pointer; background: #111827; color: white; }
    input[type="text"] { flex: 1; padding: 10px 12px; border-radius: 10px; border: 1px solid #d1d5db; }
    .done { text-decoration: line-through; color: #6b7280; }
  </style>
</head>
<body>
  <main>
    <h1>Express Todo</h1>
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
          await fetch('/api/todos/' + todo.id, {
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
          await fetch('/api/todos/' + todo.id, { method: 'DELETE' });
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
</html>`);
});

app.get("/api/todos", (_req, res) => {
  res.json(todos);
});

app.post("/api/todos", (req, res) => {
  const title = String(req.body.title || "").trim();
  const todo = { id: nextId++, title, completed: false };
  todos.push(todo);
  res.status(201).json(todo);
});

app.patch("/api/todos/:id", (req, res) => {
  const todo = todos.find((item) => item.id === Number(req.params.id));
  if (!todo) {
    res.status(404).json({ error: "Todo not found" });
    return;
  }
  todo.completed = Boolean(req.body.completed);
  res.json(todo);
});

app.delete("/api/todos/:id", (req, res) => {
  const index = todos.findIndex((item) => item.id === Number(req.params.id));
  if (index === -1) {
    res.status(404).json({ error: "Todo not found" });
    return;
  }
  todos.splice(index, 1);
  res.json({ ok: true });
});

app.listen(port, () => {
  console.log(`Express Todo listening on http://127.0.0.1:${port}`);
});
