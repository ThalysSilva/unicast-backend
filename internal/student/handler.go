package student

import "github.com/gin-gonic/gin"

type handler struct {
	service Service
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
}

func NewHandler(service Service) Handler {
	return &handler{
		service: service,
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
