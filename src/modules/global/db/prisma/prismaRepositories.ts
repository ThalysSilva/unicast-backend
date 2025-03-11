import { Provider } from '@nestjs/common';
import { UserRepository } from 'src/repositories/userRepository';
import { PrismaUserRepository } from './repositories/prismaUserRepository';

export const prismaRepositories = [
  {
    provide: UserRepository,
    useClass: PrismaUserRepository,
  },
] as Provider[];
