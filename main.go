package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Server struct {
	mu     sync.Mutex
	nextID int
	todos  []Todo
}

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Go Todo</title>
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
    <h1>Go Todo</h1>
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
</html>`))

func main() {
	server := &Server{
		nextID: 3,
		todos: []Todo{
			{ID: 1, Title: "Buy milk", Completed: false},
			{ID: 2, Title: "Write some code", Completed: true},
		},
	}

	http.HandleFunc("/", server.handleIndex)
	http.HandleFunc("/api/todos", server.handleTodos)
	http.HandleFunc("/api/todos/", server.handleTodoByID)

	log.Println("Go Todo listening on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTemplate.Execute(w, nil)
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, http.StatusOK, s.todos)
	case http.MethodPost:
		var payload struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		todo := Todo{ID: s.nextID, Title: strings.TrimSpace(payload.Title), Completed: false}
		s.nextID++
		s.todos = append(s.todos, todo)
		s.mu.Unlock()
		writeJSON(w, http.StatusCreated, todo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodoByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/todos/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var payload struct {
			Completed bool `json:"completed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		for i := range s.todos {
			if s.todos[i].ID == id {
				s.todos[i].Completed = payload.Completed
				writeJSON(w, http.StatusOK, s.todos[i])
				return
			}
		}
		http.Error(w, "Todo not found", http.StatusNotFound)
	case http.MethodDelete:
		s.mu.Lock()
		defer s.mu.Unlock()
		for i := range s.todos {
			if s.todos[i].ID == id {
				s.todos = append(s.todos[:i], s.todos[i+1:]...)
				writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
				return
			}
		}
		http.Error(w, "Todo not found", http.StatusNotFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
