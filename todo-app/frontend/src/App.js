import React, { useState, useEffect } from "react";
import "./App.css";

function App() {
  const [tasks, setTasks] = useState([]);
  const [newTask, setNewTask] = useState("");

  const API_BASE = "http://192.168.0.105:8080"; // Your backend IP

  // Fetch tasks
  const fetchTasks = () => {
    fetch(`${API_BASE}/tasks`)
      .then((res) => res.json())
      .then((data) => setTasks(data.tasks))
      .catch((err) => console.error("Error fetching tasks:", err));
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  // Add task
  const addTask = () => {
    if (!newTask.trim()) return;
    fetch(`${API_BASE}/tasks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: newTask, completed: false }),
    })
      .then((res) => res.json())
      .then((task) => {
        setTasks([...tasks, task]);
        setNewTask("");
      })
      .catch((err) => console.error("Error adding task:", err));
  };

  // Delete task
  const deleteTask = (id) => {
    fetch(`${API_BASE}/tasks/${id}`, { method: "DELETE" })
      .then(() => setTasks(tasks.filter((task) => task.id !== id)))
      .catch((err) => console.error("Error deleting task:", err));
  };

  // Toggle completion (persisted)
  const toggleCompletion = (id, completed) => {
    fetch(`${API_BASE}/tasks/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ completed: !completed }),
    })
      .then(() => {
        setTasks(
          tasks.map((task) =>
            task.id === id ? { ...task, completed: !completed } : task
          )
        );
      })
      .catch((err) => console.error("Error updating task:", err));
  };

  return (
    <div className="App">
      <h1>Todo List</h1>
      <div>
        <input
          type="text"
          placeholder="New task..."
          value={newTask}
          onChange={(e) => setNewTask(e.target.value)}
        />
        <button onClick={addTask}>Add</button>\
	  {/* 🔴 Datadog Error Trigger Button */}
  <button
    onClick={() =>
      fetch(`${API_BASE}/error-test`)
        .then((res) => res.json())
        .then((data) => console.log("Triggered error:", data))
        .catch((err) => console.error("Error calling /error-test:", err))
    }
    style={{ marginLeft: "10px", backgroundColor: "#f66", color: "#fff" }}
  >
    Trigger Error
  </button>
      </div>
      <ul>
        {tasks.map((task) => (
          <li
            key={task.id}
            style={{ display: "flex", alignItems: "center", gap: "10px", marginBottom: "5px" }}
          >
            <input
              type="checkbox"
              checked={task.completed}
              onChange={() => toggleCompletion(task.id, task.completed)}
            />
            <span style={{ textDecoration: task.completed ? "line-through" : "none" }}>
              {task.title}
            </span>
            <button onClick={() => deleteTask(task.id)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;

