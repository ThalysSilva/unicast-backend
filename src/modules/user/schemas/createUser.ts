import { createZodDto } from '@anatine/zod-nestjs';
import { z } from 'zod';

export const createUserSchema = z.object({
  name: z
    .string({
      required_error: 'Nome é obrigatório',
    })
    .min(3, 'Nome deve ter pelo menos 3 caracteres'),
  email: z
    .string({ required_error: 'Email é obrigatório' })
    .email('E-mail inválido'),
  password: z
    .string({
      required_error: 'Senha é obrigatória',
    })
    .min(8, 'Senha deve ter pelo menos 8 caracteres'),
});

export type CreateUserSchemaData = z.infer<typeof createUserSchema>;
export class CreateUserDto extends createZodDto(createUserSchema) {}
