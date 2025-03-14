import { Injectable } from '@nestjs/common';
import { UserWithPassword } from 'src/@entities/user';
import { BadRequestError } from 'src/common/applicationError';
import { UserRepository } from 'src/repositories/userRepository';
import { OmitDefaultData } from 'src/utils/types';
import * as bcrypt from 'bcrypt';

@Injectable()
export class UserService {
  constructor(private readonly userRepository: UserRepository) {}

  async create(
    user: Omit<OmitDefaultData<UserWithPassword>, 'refreshToken' | 'salt'>,
  ) {
    const salt = await bcrypt.genSalt(10);
    const password = await bcrypt.hash(user.password, salt);

    const emailExists = await this.userRepository.findByEmail(user.email);
    if (emailExists) {
      throw new BadRequestError({
        message: 'Email já está cadastrado',
        action: 'UserService.create',
        saveLog: false,
      });
    }
    return this.userRepository.create({ ...user, password, salt });
  }
}
