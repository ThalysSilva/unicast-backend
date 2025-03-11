import { Test, TestingModule } from '@nestjs/testing';
import { UserController } from './user.controller';
import { UserService } from './user.service';
import { CreateUserDto } from './schemas/createUser';
import { mock } from 'jest-mock-extended';
import { User } from 'src/@entities/user';

const userServiceMock = mock<UserService>();

describe('UserController', () => {
  let controller: UserController;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [UserController],
      providers: [{ provide: UserService, useValue: userServiceMock }],
    }).compile();

    controller = module.get<UserController>(UserController);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should create a new user', async () => {
    const user: Pick<User, 'name' | 'email'> = {
      name: 'John Doe',
      email: 'johndoe@example.com',
    };

    const userInput: CreateUserDto = {
      ...user,
      password: 'password123',
    };

    const createdUser: User = {
      ...user,
      id: '1',
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    userServiceMock.create.mockResolvedValueOnce(createdUser);

    const result = await controller.createUser(userInput);

    expect(result).toEqual(createdUser);
    expect(userServiceMock.create).toHaveBeenCalledTimes(1);
    expect(userServiceMock.create).toHaveBeenCalledWith(userInput);
  });
});
