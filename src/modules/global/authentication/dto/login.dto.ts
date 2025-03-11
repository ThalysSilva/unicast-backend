import { ApiProperty } from '@nestjs/swagger';

export class LoginData {
  @ApiProperty({
    example: 'user@email.com',
    description: 'Email do usuário',
  })
  email: string;

  @ApiProperty({
    example: 'John Doe',
    description: 'Nome do usuário',
  })
  name: string;

  @ApiProperty({
    example: '1',
    description: 'Identificador único do usuário',
  })
  id: string;
}

export class LoginResponseDto {
  @ApiProperty({
    type: LoginData, // Define corretamente o tipo para o Swagger
    description: 'Dados do usuário autenticado',
  })
  data: LoginData;

  @ApiProperty({
    example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9',
    description: 'Token de autenticação',
  })
  token: string;

  @ApiProperty({
    example: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9',
    description: 'Token de atualização',
  })
  refreshToken: string;
}
