package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"

    // Datadog tracing imports
    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
    gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

type Task struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

func main() {
    // ✅ Start Datadog tracer
    tracer.Start(
        tracer.WithEnv("todo-app"),
        tracer.WithService("todo-backend"),
        tracer.WithServiceVersion("1.0"),
    )
    defer tracer.Stop()

    // Build Postgres connection string
    dbURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_NAME"),
    )

    // Connect to Postgres
    pool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatal("Unable to connect to DB:", err)
    }
    defer pool.Close()

    // Initialize Gin with Datadog middleware
    r := gin.Default()
    r.Use(gintrace.Middleware("todo-backend"))

    // Enable CORS for frontend access
    r.Use(cors.Default())

    // 🔴 Simulate random errors for testing Datadog APM
    r.GET("/error-test", func(c *gin.Context) {
        log.Println("Simulating server error...")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Simulated internal error for Datadog test",
        })
    })

    // 🔴 Simulate DB query failure
    r.GET("/db-error", func(c *gin.Context) {
        _, err := pool.Exec(context.Background(), "SELECT * FROM non_existing_table")
        if err != nil {
            log.Println("DB query failed:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "This should not happen"})
    })

    // 🟢 GET /tasks → fetch all tasks
    r.GET("/tasks", func(c *gin.Context) {
        rows, err := pool.Query(context.Background(), "SELECT id, title, completed FROM tasks")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var tasks []Task
        for rows.Next() {
            var t Task
            if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }
            tasks = append(tasks, t)
        }

        if tasks == nil {
            tasks = []Task{}
        }

        c.JSON(http.StatusOK, gin.H{"tasks": tasks})
    })

    // 🟢 POST /tasks → add a new task
    r.POST("/tasks", func(c *gin.Context) {
        var newTask Task
        if err := c.ShouldBindJSON(&newTask); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        err := pool.QueryRow(context.Background(),
            "INSERT INTO tasks (title, completed) VALUES ($1, $2) RETURNING id",
            newTask.Title, newTask.Completed).Scan(&newTask.ID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, newTask)
    })

    // 🟢 DELETE /tasks/:id → delete a task
    r.DELETE("/tasks/:id", func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
            return
        }

        cmdTag, err := pool.Exec(context.Background(), "DELETE FROM tasks WHERE id=$1", id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        if cmdTag.RowsAffected() == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
    })

    // 🟢 PATCH /tasks/:id → toggle completed
    r.PATCH("/tasks/:id", func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
            return
        }

        var task Task
        if err := c.ShouldBindJSON(&task); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        cmdTag, err := pool.Exec(context.Background(),
            "UPDATE tasks SET completed=$1 WHERE id=$2",
            task.Completed, id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        if cmdTag.RowsAffected() == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
    })

    // 🟢 Start server
    r.Run(":8080")
}

