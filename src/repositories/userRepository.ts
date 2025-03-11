import { OmitDefaultData } from 'src/utils/types';
import { Repository } from './repository';
import { User, UserWithPassword } from 'src/@entities/user';

export abstract class UserRepository extends Repository {
  abstract findByIdWithPassword(user: any): Promise<UserWithPassword | null>;
  abstract findById(id: string): Promise<User | null>;
  abstract update(
    id: string,
    data: Partial<OmitDefaultData<UserWithPassword>>,
  ): Promise<User | null>;
  abstract create(
    user: Omit<OmitDefaultData<UserWithPassword>, 'refreshToken'>,
  ): Promise<User | null>;
  abstract findByEmail(email: string): Promise<User | null>;
}
