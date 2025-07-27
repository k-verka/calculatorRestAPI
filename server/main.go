// Импорты
package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"database/sql"
    _ "github.com/lib/pq"
)
//GLOBAL
var db *sql.DB
// Модели
type Calculation struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
	Result     string `json:"result"`
}

type CalculationRequest struct {
	Expression string `json:"expression"`
}

var calculations = []Calculation{}

// Логика вычисления
func calculateExpression(expression string) (string, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return "", err
	}
	result, err := expr.Evaluate(nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", result), err
}
//Методы
// GET
func getCalculations(c echo.Context) error {
	return c.JSON(http.StatusOK, calculations)
}

// POST
func postCalculations(c echo.Context) error {
	req := new(CalculationRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid POST request"})
	}

	result, err := calculateExpression(req.Expression)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid expression"})
	}

	_, err = db.Exec(`INSERT INTO expressions (id, expression, result) VALUES ($1, $2, $3)`, c.Param("id"), req.Expression, result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB insert failed"})
	}

	calc := Calculation{
		ID:         uuid.NewString(),
		Expression: req.Expression,
		Result:     result,
	}
	calculations = append(calculations, calc)
	return c.JSON(http.StatusCreated, calc)
}

// PATCH
func patchCalculation(c echo.Context) error {
	id := c.Param("id")
	req := new(CalculationRequest)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid PATCH request"})
	}

	result, err := calculateExpression(req.Expression)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid expression"})
	}

	for i := range calculations {
		if calculations[i].ID == id {
			calculations[i].Expression = req.Expression
			calculations[i].Result = result
			return c.JSON(http.StatusOK, calculations[i])
		}
	}

	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Calculation not found"})

}

// DELETE
func deleteCalculation(c echo.Context) error {
	id := c.Param("id")

	for i, calculation := range calculations {
		if calculation.ID == id {
			calculations = slices.Delete(calculations, i, i+1)
			return c.NoContent(http.StatusNoContent)
		}
	}
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Calculation not found"})
}

// Точка входа
func main() {
	var err error
	db, err = sql.Open("postgres", "host=db user=postgres password=super0secret0cat0meow dbname=example sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	e.GET("/calculations", getCalculations)
	e.POST("/calculations", postCalculations)
	e.PATCH("/calculations/:id", patchCalculation)
	e.DELETE("/calculations/:id", deleteCalculation)

	e.Start("localhost:8080")
}
