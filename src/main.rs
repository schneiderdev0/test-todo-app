use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{Html, IntoResponse},
    routing::{get, patch},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::{
    net::SocketAddr,
    sync::{Arc, Mutex},
};

#[derive(Clone)]
struct AppState {
    todos: Arc<Mutex<Vec<Todo>>>,
    next_id: Arc<Mutex<u64>>,
}

#[derive(Clone, Serialize)]
struct Todo {
    id: u64,
    title: String,
    completed: bool,
}

#[derive(Deserialize)]
struct CreateTodo {
    title: String,
}

#[derive(Deserialize)]
struct UpdateTodo {
    completed: bool,
}

#[tokio::main]
async fn main() {
    let state = AppState {
        todos: Arc::new(Mutex::new(vec![
            Todo {
                id: 1,
                title: "Buy milk".into(),
                completed: false,
            },
            Todo {
                id: 2,
                title: "Write some code".into(),
                completed: true,
            },
        ])),
        next_id: Arc::new(Mutex::new(3)),
    };

    let app = Router::new()
        .route("/", get(index))
        .route("/api/todos", get(list_todos).post(create_todo))
        .route("/api/todos/{id}", patch(update_todo).delete(delete_todo))
        .with_state(state);

    let addr = SocketAddr::from(([127, 0, 0, 1], 4000));
    println!("Rust Todo listening on http://{addr}");

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn index() -> Html<&'static str> {
    Html(
        r#"<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Rust Todo</title>
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
    <h1>Rust Todo</h1>
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
</html>"#,
    )
}

async fn list_todos(State(state): State<AppState>) -> Json<Vec<Todo>> {
    let todos = state.todos.lock().unwrap().clone();
    Json(todos)
}

async fn create_todo(
    State(state): State<AppState>,
    Json(payload): Json<CreateTodo>,
) -> impl IntoResponse {
    let mut next_id = state.next_id.lock().unwrap();
    let todo = Todo {
        id: *next_id,
        title: payload.title.trim().to_string(),
        completed: false,
    };
    *next_id += 1;

    let mut todos = state.todos.lock().unwrap();
    todos.push(todo.clone());

    (StatusCode::CREATED, Json(todo))
}

async fn update_todo(
    Path(id): Path<u64>,
    State(state): State<AppState>,
    Json(payload): Json<UpdateTodo>,
) -> impl IntoResponse {
    let mut todos = state.todos.lock().unwrap();
    if let Some(todo) = todos.iter_mut().find(|todo| todo.id == id) {
        todo.completed = payload.completed;
        return (StatusCode::OK, Json(todo.clone())).into_response();
    }

    (StatusCode::NOT_FOUND, "Todo not found").into_response()
}

async fn delete_todo(Path(id): Path<u64>, State(state): State<AppState>) -> impl IntoResponse {
    let mut todos = state.todos.lock().unwrap();
    if let Some(index) = todos.iter().position(|todo| todo.id == id) {
        todos.remove(index);
        return (StatusCode::OK, Json(serde_json::json!({ "ok": true }))).into_response();
    }

    (StatusCode::NOT_FOUND, "Todo not found").into_response()
}
