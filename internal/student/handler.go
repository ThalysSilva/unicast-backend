package student

import (
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service       Service
	importService ImportService
}

type createStudentInput struct {
	StudentID  string        `json:"studentId" binding:"required"`
	Name       *string       `json:"name"`
	Phone      *string       `json:"phone" `
	Email      *string       `json:"email" binding:"email"`
	Annotation *string       `json:"annotation"`
	Status     StudentStatus `json:"status" binding:"required,oneof=ACTIVE CANCELED GRADUATED LOCKED"`
}

type Handler interface {
	Create() gin.HandlerFunc
	GetStudent() gin.HandlerFunc
	GetStudents() gin.HandlerFunc
	Update() gin.HandlerFunc
	Delete() gin.HandlerFunc
	ImportForCourse() gin.HandlerFunc
}

func NewHandler(service Service, importService ImportService) Handler {
	return &handler{
		service:       service,
		importService: importService,
	}
}
func (h *handler) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input createStudentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		err := h.service.Create(c.Request.Context(), input.StudentID, input.Name, input.Phone, input.Email, input.Annotation, input.Status)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Aluno criado com sucesso"})
	}
}

func (h *handler) GetStudent() gin.HandlerFunc {
	return func(c *gin.Context) {
		studentID := c.Param("id")

		student, err := h.service.GetStudent(c.Request.Context(), studentID)
		if err != nil {
			c.Error(err)
			return
		}
		if student == nil {
			c.JSON(404, gin.H{"message": "Aluno não encontrado"})
			return
		}
		c.JSON(200, student)
	}
}

func (h *handler) GetStudents() gin.HandlerFunc {
	return func(c *gin.Context) {
		program := c.Query("program")
		campus := c.Query("campus")
		course := c.Query("course")
		user := c.Query("user")
		// Filtro por disciplina, campus, cursos, usuário.
		filters := make(map[string]string)
		if program != "" {
			filters["program"] = program
		}
		if campus != "" {
			filters["campus"] = campus
		}
		if course != "" {
			filters["course"] = course
		}
		if user != "" {
			filters["user"] = user
		}

		students, err := h.service.GetStudents(c.Request.Context(), filters)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, students)
	}
}

func (h *handler) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		studentID := c.Param("id")
		var input createStudentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}

		fields := make(map[string]any)
		if input.Name != nil {
			fields["name"] = input.Name
		}
		if input.Phone != nil {
			fields["phone"] = input.Phone
		}
		if input.Email != nil {
			fields["email"] = input.Email
		}
		if input.Annotation != nil {
			fields["annotation"] = input.Annotation
		}

		if input.Status != "" {
			fields["status"] = input.Status
		}

		err := h.service.Update(c.Request.Context(), studentID, fields)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Aluno atualizado com sucesso"})
	}
}

func (h *handler) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		studentID := c.Param("id")
		err := h.service.Delete(c.Request.Context(), studentID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, gin.H{"message": "Aluno deletado com sucesso"})
	}
}

func (h *handler) ImportForCourse() gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.Param("courseId")
		modeParam := strings.ToLower(c.DefaultQuery("mode", string(ImportModeUpsert)))
		mode := ImportMode(modeParam)
		if mode != ImportModeClean && mode != ImportModeUpsert {
			c.JSON(http.StatusBadRequest, gin.H{"message": "mode inválido, use clean ou upsert"})
			return
		}

		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "arquivo 'file' é obrigatório"})
			return
		}
		defer file.Close()

		records, err := parseImportCSV(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		result, err := h.importService.ImportForCourse(c.Request.Context(), courseID, mode, records)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func parseImportCSV(file multipart.File) ([]ImportRecord, error) {
	rows, err := readCSV(file)
	if err != nil {
		return nil, err
	}

	columns, err := mapColumns(rows[0])
	if err != nil {
		return nil, err
	}

	records := make([]ImportRecord, 0, len(rows)-1)
	for i, row := range rows[1:] {
		rec, err := buildImportRecord(row, columns, i+2)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}

	return records, nil
}

func readCSV(file multipart.File) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler CSV: %w", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("arquivo precisa ter cabeçalho e pelo menos uma linha de dados")
	}
	return rows, nil
}

func mapColumns(header []string) (map[string]int, error) {
	columns := make(map[string]int, len(header))
	for i, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		if key != "" {
			columns[key] = i
		}
	}
	if _, ok := columns["studentid"]; !ok {
		return nil, fmt.Errorf("coluna studentId é obrigatória")
	}
	return columns, nil
}

func buildImportRecord(row []string, columns map[string]int, line int) (ImportRecord, error) {
	get := func(key string) string {
		idx, ok := columns[key]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	toPtr := func(value string) *string {
		if value == "" {
			return nil
		}
		return &value
	}

	status, err := parseStatus(get("status"))
	if err != nil {
		return ImportRecord{}, fmt.Errorf("linha %d: %v", line, err)
	}

	return ImportRecord{
		StudentID: get("studentid"),
		Name:      toPtr(get("name")),
		Phone:     toPtr(get("phone")),
		Email:     toPtr(get("email")),
		Status:    status,
	}, nil
}

func parseStatus(input string) (StudentStatus, error) {
	value := strings.TrimSpace(strings.ToUpper(input))
	switch value {
	case "1", "ACTIVE":
		return StudentStatusActive, nil
	case "2", "LOCKED", "TRANCADO":
		return StudentStatusLocked, nil
	case "3", "GRADUATED", "CONCLUIDO":
		return StudentStatusGraduated, nil
	case "4", "CANCELED", "CANCELADO":
		return StudentStatusCanceled, nil
	case "", "5", "PENDING", "PENDENTE":
		return StudentStatusPending, nil
	default:
		if _, err := strconv.Atoi(value); err == nil {
			return "", fmt.Errorf("status numérico %s não suportado", value)
		}
		return "", fmt.Errorf("status inválido: %s", input)
	}
}
