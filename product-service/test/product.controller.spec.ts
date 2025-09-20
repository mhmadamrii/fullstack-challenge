import { Test, TestingModule } from '@nestjs/testing';
import { ProductController } from '../src/modules/product/product.controller';
import { ProductService } from '../src/modules/product/product.service';
import { CreateProductDto } from '../src/modules/product/product.dto';

describe('ProductController', () => {
  let controller: ProductController;
  let service: ProductService;

  const mockProductService = {
    createProduct: jest.fn(),
    getAllProducts: jest.fn(),
    getProductById: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [ProductController],
      providers: [
        {
          provide: ProductService,
          useValue: mockProductService,
        },
      ],
    }).compile();

    controller = module.get<ProductController>(ProductController);
    service = module.get<ProductService>(ProductService);
  });

  it('should be defined', () => {
    expect(controller).toBeDefined();
  });

  describe('createProduct', () => {
    it('should create a product', async () => {
      const createProductDto: CreateProductDto = {
        name: 'Test Product',
        price: 100,
        qty: 10,
      };
      const expectedProduct = { id: '1', ...createProductDto };

      mockProductService.createProduct.mockResolvedValue(expectedProduct);

      const result = await controller.createProduct(createProductDto);

      expect(result).toEqual(expectedProduct);
      expect(service.createProduct).toHaveBeenCalledWith(createProductDto);
    });
  });

  describe('getProducts', () => {
    it('should return an array of products', async () => {
      const expectedProducts = [
        {
          id: '1',
          name: 'Test Product 1',
          description: 'Test Description 1',
          price: 100,
          qty: 10,
        },
        {
          id: '2',
          name: 'Test Product 2',
          description: 'Test Description 2',
          price: 200,
          qty: 20,
        },
      ];

      mockProductService.getAllProducts.mockResolvedValue(expectedProducts);

      const result = await controller.getProducts();

      expect(result).toEqual(expectedProducts);
      expect(service.getAllProducts).toHaveBeenCalled();
    });
  });

  describe('getProductById', () => {
    it('should return a single product', async () => {
      const expectedProduct = {
        id: '1',
        name: 'Test Product 1',
        description: 'Test Description 1',
        price: 100,
        qty: 10,
      };

      mockProductService.getProductById.mockResolvedValue(expectedProduct);

      const result = await controller.getProductById('1');

      expect(result).toEqual(expectedProduct);
      expect(service.getProductById).toHaveBeenCalledWith('1');
    });
  });
});
