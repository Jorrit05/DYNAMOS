import PyPDF2
import os

def split_pdf_into_pages(pdf_path, output_folder):
    # Ensure the output folder exists
    if not os.path.exists(output_folder):
        os.makedirs(output_folder)
    name= pdf_path.split('/')[-1].split('.')[0]
    # Open the source PDF
    with open(pdf_path, 'rb') as infile:
        reader = PyPDF2.PdfReader(infile)
        num_pages = len(reader.pages)

        # Iterate through all the pages
        for i in range(num_pages):
            writer = PyPDF2.PdfWriter()
            writer.add_page(reader.pages[i])

            # Output each page as a separate PDF
            output_filename = os.path.join(output_folder, f'{name}_{i+1}.pdf')
            with open(output_filename, 'wb') as outfile:
                writer.write(outfile)
            print(f"Created: {output_filename}")

# Usage
# pdf_path = '/Users/jorrit/Documents/uva/DYNAMOS/docs/slides/Presentation_ofc.pdf'
pdf_path = './slides_demo_movie.pdf'
output_folder = '/Users/jorrit/Documents/uva/DYNAMOS/docs/slides/separate'
split_pdf_into_pages(pdf_path, output_folder)
