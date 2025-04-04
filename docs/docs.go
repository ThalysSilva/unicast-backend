// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Gera o acesso a um usuário no sistema",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Gera o acesso a um usuário",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.LoginInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.DefaultResponse-github_com_ThalysSilva_unicast-backend_internal_services_LoginResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "description": "Remove o acesso a um usuário do sistema",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Remove o acesso a um usuário",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bearer token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "message": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Atualiza o Refresh Token do usuário no sistema",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Atualiza o Refresh Token do usuário",
                "parameters": [
                    {
                        "description": "Refresh token",
                        "name": "refreshToken",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.RefreshInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_services.RefreshResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Registra um novo usuário no sistema",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Registra um novo usuário",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.RegisterInput"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/message/send": {
            "post": {
                "description": "Envia uma mensagem via email e WhatsApp",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "message"
                ],
                "summary": "Envia uma mensagem",
                "parameters": [
                    {
                        "description": "Message data",
                        "name": "message",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/internal_handlers.MessageInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.DefaultResponse-internal_handlers_MessageDataResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_ThalysSilva_unicast-backend_internal_models.Attachment": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "fileName": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_models.DefaultResponse-github_com_ThalysSilva_unicast-backend_internal_services_LoginResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_services.LoginResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_models.DefaultResponse-internal_handlers_MessageDataResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/internal_handlers.MessageDataResponse"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_models.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_models_entities.Student": {
            "type": "object",
            "properties": {
                "annotation": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "status": {
                    "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models_entities.StudentStatus"
                },
                "studentId": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_models_entities.StudentStatus": {
            "type": "string",
            "enum": [
                "ACTIVE",
                "CANCELED",
                "GRADUATED",
                "LOCKED"
            ],
            "x-enum-varnames": [
                "StudentStatusActive",
                "StudentStatusCanceled",
                "StudentStatusGraduated",
                "StudentStatusLocked"
            ]
        },
        "github_com_ThalysSilva_unicast-backend_internal_models_entities.User": {
            "type": "object",
            "required": [
                "email",
                "name"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_services.LoginResponse": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "jwe": {
                    "type": "string"
                },
                "refreshToken": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models_entities.User"
                }
            }
        },
        "github_com_ThalysSilva_unicast-backend_internal_services.RefreshResponse": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "refreshToken": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models_entities.User"
                }
            }
        },
        "internal_handlers.LoginInput": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "internal_handlers.MessageDataResponse": {
            "type": "object",
            "properties": {
                "emailsFailed": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models_entities.Student"
                    }
                },
                "whatsappFailed": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models_entities.Student"
                    }
                }
            }
        },
        "internal_handlers.MessageInput": {
            "type": "object",
            "required": [
                "body",
                "from",
                "jwe",
                "smtp_id",
                "subject",
                "to",
                "whatsapp_id"
            ],
            "properties": {
                "attachment": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_ThalysSilva_unicast-backend_internal_models.Attachment"
                    }
                },
                "body": {
                    "type": "string"
                },
                "from": {
                    "type": "string"
                },
                "jwe": {
                    "type": "string"
                },
                "smtp_id": {
                    "type": "string"
                },
                "subject": {
                    "type": "string"
                },
                "to": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "whatsapp_id": {
                    "type": "string"
                }
            }
        },
        "internal_handlers.RefreshInput": {
            "type": "object",
            "required": [
                "refreshToken"
            ],
            "properties": {
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "internal_handlers.RegisterInput": {
            "type": "object",
            "required": [
                "email",
                "name",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
